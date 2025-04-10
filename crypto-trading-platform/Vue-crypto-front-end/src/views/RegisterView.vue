<template>
  <div class="container">
    <h2>注册</h2>
    <input v-model="username" placeholder="用户名" />
    <input v-model="password" type="password" placeholder="密码" />
    <input v-model="email" type="email" placeholder="邮箱" />
    <button @click="register">注册</button>
    <p v-if="error">{{ error }}</p>
    <p>已有账号？<router-link to="/">去登录</router-link></p>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue';
import { useRouter } from 'vue-router';

export default defineComponent({
  name: 'RegisterView',
  setup() {
    const username = ref('');
    const password = ref('');
    const email = ref('');
    const error = ref('');
    const router = useRouter();

    const register = async () => {
      // 前端验证
      if (!username.value || !password.value || !email.value) {
        error.value = '用户名、密码和邮箱不能为空';
        return;
      }

      // 简单的邮箱格式验证
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(email.value)) {
        error.value = '请输入有效的邮箱地址';
        return;
      }

      try {
        const response = await fetch('http://localhost:8888/api/user/register', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json',
          },
          body: JSON.stringify({
            username: username.value,
            password: password.value,
            email: email.value
          }),
        });

        const data = await response.json();

        if (response.ok && data.code === 0) {
          alert('注册成功！请登录。');
          router.push('/');
          // 清空表单
          username.value = '';
          password.value = '';
          email.value = '';
          error.value = '';
        } else {
          error.value = data.msg || '注册失败，请稍后重试';
        }
      } catch (err) {
        error.value = '网络错误：' + (err as Error).message;
      }
    };

    return { username, password, email, error, register };
  },
});
</script>

<style scoped>
.container {
  max-width: 400px;
  margin: 50px auto;
  padding: 20px;
  border: 1px solid #ccc;
  border-radius: 5px;
  text-align: center;
}
input {
  display: block;
  width: 100%;
  padding: 10px;
  margin: 10px 0;
  border: 1px solid #ccc;
  border-radius: 5px;
}
button {
  padding: 10px 20px;
  background-color: #4caf50;
  color: white;
  border: none;
  border-radius: 5px;
  cursor: pointer;
}
button:hover {
  background-color: #45a049;
}
a {
  color: #4caf50;
  text-decoration: none;
}
a:hover {
  text-decoration: underline;
}
</style>