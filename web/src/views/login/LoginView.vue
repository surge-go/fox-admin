<script setup lang="ts">
import { LockClosedOutline, LogoGithub, LogoWechat, PersonOutline, ShieldCheckmarkOutline } from '@vicons/ionicons5'
import { useMessage } from 'naive-ui'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const authStore = useAuthStore()

const form = reactive({
  username: 'admin',
  password: 'admin123',
  remember: true,
})

function handleLogin() {
  authStore.setSession('fox-admin-demo-token', form.username)
  message.success('登录成功')
  router.replace((route.query.redirect as string) || '/dashboard')
}
</script>

<template>
  <main class="login-page">
    <section class="login-panel">
      <header class="login-brand">
        <div class="login-brand__logo">
          <n-icon :component="ShieldCheckmarkOutline" />
        </div>
        <span>Fox Admin</span>
      </header>

      <p class="login-subtitle">Fox Admin 中后台管理系统</p>

      <div class="login-heading">
        <h1>账号登录</h1>
        <span />
        <p>欢迎回来，请登录您的账号</p>
      </div>

      <n-form class="login-form" :show-label="false" @submit.prevent="handleLogin">
        <n-form-item>
          <n-input v-model:value="form.username" size="large" placeholder="请输入账号">
            <template #prefix>
              <n-icon :component="PersonOutline" />
            </template>
          </n-input>
        </n-form-item>

        <n-form-item>
          <n-input
            v-model:value="form.password"
            size="large"
            type="password"
            placeholder="请输入密码"
            show-password-on="click"
          >
            <template #prefix>
              <n-icon :component="LockClosedOutline" />
            </template>
          </n-input>
        </n-form-item>

        <div class="login-options">
          <n-checkbox v-model:checked="form.remember">自动登录</n-checkbox>
          <n-button text type="primary">忘记密码</n-button>
        </div>

        <n-button attr-type="submit" class="login-submit" type="primary" size="large" block>
          登录
        </n-button>
      </n-form>

      <footer class="login-footer">
        <span>其它登录方式</span>
        <n-button circle tertiary>
          <template #icon>
            <n-icon :component="LogoGithub" />
          </template>
        </n-button>
        <n-button circle tertiary>
          <template #icon>
            <n-icon :component="LogoWechat" />
          </template>
        </n-button>
        <n-button text type="primary" class="login-register">注册账号</n-button>
      </footer>
    </section>
  </main>
</template>

<style scoped>
.login-page {
  position: relative;
  overflow: hidden;
  min-height: 100vh;
  padding: 32px 20px;
  background:
    linear-gradient(90deg, rgba(15, 23, 42, 0.2), rgba(15, 23, 42, 0.03)),
    url('https://img.netbian.com/file/2026/0524/221815swejN.jpg') center / cover no-repeat;
}

.login-page::before {
  position: absolute;
  inset: 0;
  content: '';
  background: rgba(255, 255, 255, 0.18);
}

.login-panel {
  position: relative;
  z-index: 1;
  width: min(420px, 100%);
  padding: 34px 34px 30px;
  border: 1px solid rgba(255, 255, 255, 0.68);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.9);
  box-shadow: 0 18px 56px rgba(31, 41, 55, 0.2);
  backdrop-filter: blur(10px);
}

.login-brand {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: #1f3658;
  font-size: 24px;
  font-weight: 650;
}

.login-brand__logo {
  display: grid;
  width: 32px;
  height: 32px;
  place-items: center;
  border-radius: 7px;
  background: linear-gradient(135deg, #1d4ed8, #3b82f6);
  color: #fff;
  font-size: 21px;
  box-shadow: 0 8px 18px rgba(37, 99, 235, 0.24);
}

.login-subtitle {
  margin-top: 16px;
  color: #52637a;
  text-align: center;
  font-size: 14px;
}

.login-heading {
  margin-top: 12px;
  text-align: center;
}

.login-heading h1 {
  color: #1f2937;
  font-size: 22px;
  font-weight: 700;
}

.login-heading span {
  display: block;
  width: 36px;
  height: 2px;
  margin: 8px auto 10px;
  border-radius: 999px;
  background: #2563eb;
}

.login-heading p {
  color: #697586;
  font-size: 13px;
}

.login-form {
  margin-top: 24px;
}

.login-form :deep(.n-form-item) {
  margin-bottom: 16px;
}

.login-form :deep(.n-input) {
  --n-border-radius: 6px;
  --n-height: 42px;
  --n-font-size: 14px;
}

.login-options {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 2px 0 22px;
  font-size: 13px;
}

.login-submit {
  height: 44px;
  font-size: 15px;
  font-weight: 600;
  box-shadow: 0 10px 24px rgba(37, 99, 235, 0.22);
}

.login-footer {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 20px;
  color: #4b5563;
  font-size: 13px;
}

.login-register {
  margin-left: auto;
}

@media (max-width: 640px) {
  .login-panel {
    padding: 30px 22px 26px;
  }

  .login-brand {
    font-size: 22px;
  }

  .login-footer {
    flex-wrap: wrap;
  }

  .login-register {
    margin-left: 0;
  }
}
</style>
