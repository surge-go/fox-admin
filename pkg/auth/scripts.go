package auth

const issueScript = `
local subject_z = KEYS[1]
local platform_z = KEYS[2]
local device_z = KEYS[3]

local prefix = ARGV[1]
local session_id = ARGV[2]
local session_json = ARGV[3]
local session_ttl = tonumber(ARGV[4])
local refresh_hash = ARGV[5]
local refresh_ttl = tonumber(ARGV[6])
local issued_score = tonumber(ARGV[7])
local policy_max = tonumber(ARGV[8])
local platform_max = tonumber(ARGV[9])
local strategy = ARGV[10]
local subject_type = ARGV[11]
local subject_id = ARGV[12]
local platform = ARGV[13]
local device_id = ARGV[14]
local index_ttl = tonumber(ARGV[15])
local exclusive_count = tonumber(ARGV[16])

local function session_key(sid) return prefix .. ":session:" .. sid end
local function meta_key(sid) return prefix .. ":session_meta:" .. sid end
local function session_refresh_key(sid) return prefix .. ":session_refresh:" .. sid end
local function refresh_key(hash) return prefix .. ":refresh:" .. hash end
local function platform_z_key(st, si, pf) return prefix .. ":platform_sessions:" .. st .. ":" .. si .. ":" .. pf end
local function device_z_key(st, si, did) return prefix .. ":device_sessions:" .. st .. ":" .. si .. ":" .. did end

local function cleanup_zset(zkey)
  local members = redis.call("ZRANGE", zkey, 0, -1)
  for _, sid in ipairs(members) do
    if redis.call("EXISTS", session_key(sid)) == 0 then
      redis.call("ZREM", zkey, sid)
    end
  end
end

cleanup_zset(subject_z)
cleanup_zset(platform_z)
cleanup_zset(device_z)
for i = 1, exclusive_count do
  cleanup_zset(KEYS[3 + i])
end

local conflicts = {}
local seen = {}
local function add_conflict(sid)
  if sid ~= session_id and seen[sid] == nil then
    seen[sid] = true
    table.insert(conflicts, sid)
  end
end

local function add_all(zkey)
  local members = redis.call("ZRANGE", zkey, 0, -1)
  for _, sid in ipairs(members) do add_conflict(sid) end
end

local function add_over_limit(zkey, max_sessions)
  if max_sessions <= 0 then return end
  local count = redis.call("ZCARD", zkey)
  if count < max_sessions then return end
  local take = count - max_sessions + 1
  local members
  if strategy == "all" then
    members = redis.call("ZRANGE", zkey, 0, -1)
  else
    members = redis.call("ZRANGE", zkey, 0, take - 1)
  end
  for _, sid in ipairs(members) do add_conflict(sid) end
end

for i = 1, exclusive_count do
  add_all(KEYS[3 + i])
end
add_over_limit(platform_z, platform_max)
add_over_limit(subject_z, policy_max)

if strategy == "latest" and #conflicts > 0 then
  return {0, "login_conflict"}
end

local revoked = {}
for _, sid in ipairs(conflicts) do
  local old_session = redis.call("GET", session_key(sid))
  if old_session then table.insert(revoked, old_session) end
  local old_hash = redis.call("GET", session_refresh_key(sid))
  if old_hash then redis.call("DEL", refresh_key(old_hash)) end
  local old_subject_type = redis.call("HGET", meta_key(sid), "subject_type")
  local old_subject_id = redis.call("HGET", meta_key(sid), "subject_id")
  local old_platform = redis.call("HGET", meta_key(sid), "platform")
  local old_device_id = redis.call("HGET", meta_key(sid), "device_id")
  if old_subject_type and old_subject_id and old_platform then
    redis.call("ZREM", platform_z_key(old_subject_type, old_subject_id, old_platform), sid)
  end
  if old_subject_type and old_subject_id and old_device_id then
    redis.call("ZREM", device_z_key(old_subject_type, old_subject_id, old_device_id), sid)
  end
  redis.call("ZREM", subject_z, sid)
  redis.call("DEL", session_key(sid), meta_key(sid), session_refresh_key(sid))
end

redis.call("SET", session_key(session_id), session_json, "EX", session_ttl)
redis.call("HSET", meta_key(session_id), "subject_type", subject_type, "subject_id", subject_id, "platform", platform, "device_id", device_id)
redis.call("EXPIRE", meta_key(session_id), session_ttl)
redis.call("SET", refresh_key(refresh_hash), session_id, "EX", refresh_ttl)
redis.call("SET", session_refresh_key(session_id), refresh_hash, "EX", session_ttl)
redis.call("ZADD", subject_z, issued_score, session_id)
redis.call("ZADD", platform_z, issued_score, session_id)
redis.call("ZADD", device_z, issued_score, session_id)
redis.call("EXPIRE", subject_z, index_ttl)
redis.call("EXPIRE", platform_z, index_ttl)
redis.call("EXPIRE", device_z, index_ttl)

local result = {1}
for _, item in ipairs(revoked) do table.insert(result, item) end
return result
`

const refreshScript = `
local old_refresh_key = KEYS[1]
local new_refresh_key = KEYS[2]
local reuse_key = KEYS[3]
local session_key = KEYS[4]
local session_meta_key = KEYS[5]
local session_refresh_key = KEYS[6]
local subject_z = KEYS[7]
local platform_z = KEYS[8]
local device_z = KEYS[9]

local old_hash = ARGV[1]
local new_hash = ARGV[2]
local session_id = ARGV[3]
local session_json = ARGV[4]
local session_ttl = tonumber(ARGV[5])
local refresh_ttl = tonumber(ARGV[6])
local reuse_ttl = tonumber(ARGV[7])
local rotation = ARGV[8]
local index_ttl = tonumber(ARGV[9])

local stored_session_id = redis.call("GET", old_refresh_key)
if not stored_session_id then
  if redis.call("EXISTS", reuse_key) == 1 then
    return {0, "reused"}
  end
  return {0, "invalid"}
end
if stored_session_id ~= session_id then
  return {0, "invalid"}
end
if redis.call("EXISTS", session_key) == 0 then
  return {0, "invalid"}
end
local current_hash = redis.call("GET", session_refresh_key)
if current_hash ~= old_hash then
  return {0, "invalid"}
end

-- SETNX on the reuse key is the rotation-independent replay guard.
-- If the key already exists, this refresh is the second (or later) use
-- of the same hash, regardless of whether RefreshRotation is on or off.
-- The rotation ARGV is now historical at the Lua level; it is kept only
-- for wire compatibility with the Go caller signature.
local reuse_set = redis.call("SETNX", reuse_key, session_id)
if reuse_set == 0 then
  return {0, "reused"}
end
redis.call("EXPIRE", reuse_key, reuse_ttl)

-- Always rotate the refresh slot. When RefreshRotation is off, Go sets
-- newRefreshToken == refreshToken and newHash == oldHash, so
-- new_refresh_key == old_refresh_key. The DEL below removes the slot,
-- then the SET below recreates it with the same value/TTL — net effect is
-- a no-op on Redis state, but it lets us reuse the same Lua for both modes.
redis.call("DEL", old_refresh_key)
redis.call("SET", new_refresh_key, session_id, "EX", refresh_ttl)
redis.call("SET", session_refresh_key, new_hash, "EX", session_ttl)
redis.call("SET", session_key, session_json, "EX", session_ttl)
redis.call("EXPIRE", session_meta_key, session_ttl)
redis.call("EXPIRE", subject_z, index_ttl)
redis.call("EXPIRE", platform_z, index_ttl)
redis.call("EXPIRE", device_z, index_ttl)
return {1}
`

const revokeSessionScript = `
local prefix = ARGV[1]
local session_id = ARGV[2]

local function session_key(sid) return prefix .. ":session:" .. sid end
local function meta_key(sid) return prefix .. ":session_meta:" .. sid end
local function session_refresh_key(sid) return prefix .. ":session_refresh:" .. sid end
local function refresh_key(hash) return prefix .. ":refresh:" .. hash end
local function subject_z_key(st, si) return prefix .. ":subject_sessions:" .. st .. ":" .. si end
local function platform_z_key(st, si, pf) return prefix .. ":platform_sessions:" .. st .. ":" .. si .. ":" .. pf end
local function device_z_key(st, si, did) return prefix .. ":device_sessions:" .. st .. ":" .. si .. ":" .. did end

local session_json = redis.call("GET", session_key(session_id))
if not session_json then return {0, "not_found"} end
local refresh_hash = redis.call("GET", session_refresh_key(session_id))
if refresh_hash then redis.call("DEL", refresh_key(refresh_hash)) end
local subject_type = redis.call("HGET", meta_key(session_id), "subject_type")
local subject_id = redis.call("HGET", meta_key(session_id), "subject_id")
local platform = redis.call("HGET", meta_key(session_id), "platform")
local device_id = redis.call("HGET", meta_key(session_id), "device_id")
if subject_type and subject_id then
  redis.call("ZREM", subject_z_key(subject_type, subject_id), session_id)
end
if subject_type and subject_id and platform then
  redis.call("ZREM", platform_z_key(subject_type, subject_id, platform), session_id)
end
if subject_type and subject_id and device_id then
  redis.call("ZREM", device_z_key(subject_type, subject_id, device_id), session_id)
end
redis.call("DEL", session_key(session_id), meta_key(session_id), session_refresh_key(session_id))
return {1, session_json}
`

const revokeSubjectScript = `
local prefix = ARGV[1]
local subject_type = ARGV[2]
local subject_id = ARGV[3]
local subject_z = prefix .. ":subject_sessions:" .. subject_type .. ":" .. subject_id

local function session_key(sid) return prefix .. ":session:" .. sid end
local function meta_key(sid) return prefix .. ":session_meta:" .. sid end
local function session_refresh_key(sid) return prefix .. ":session_refresh:" .. sid end
local function refresh_key(hash) return prefix .. ":refresh:" .. hash end
local function platform_z_key(st, si, pf) return prefix .. ":platform_sessions:" .. st .. ":" .. si .. ":" .. pf end
local function device_z_key(st, si, did) return prefix .. ":device_sessions:" .. st .. ":" .. si .. ":" .. did end

local sessions = redis.call("ZRANGE", subject_z, 0, -1)
local revoked = {}
for _, sid in ipairs(sessions) do
  local session_json = redis.call("GET", session_key(sid))
  if session_json then table.insert(revoked, session_json) end
  local refresh_hash = redis.call("GET", session_refresh_key(sid))
  if refresh_hash then redis.call("DEL", refresh_key(refresh_hash)) end
  local platform = redis.call("HGET", meta_key(sid), "platform")
  local device_id = redis.call("HGET", meta_key(sid), "device_id")
  if platform then redis.call("ZREM", platform_z_key(subject_type, subject_id, platform), sid) end
  if device_id then redis.call("ZREM", device_z_key(subject_type, subject_id, device_id), sid) end
  redis.call("DEL", session_key(sid), meta_key(sid), session_refresh_key(sid))
end
redis.call("DEL", subject_z)
local result = {1}
for _, item in ipairs(revoked) do table.insert(result, item) end
return result
`
