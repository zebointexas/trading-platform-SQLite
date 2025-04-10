<template>
  <div class="container">
    <h2>登录</h2>
    <input v-model="username" placeholder="用户名" />
    <input v-model="password" type="password" placeholder="密码" />
    <button @click="login">登录</button>
    <p v-if="error">{{ error }}</p>
    <p>没有账号？<router-link to="/register">去注册</router-link></p> <!-- 新增 -->
  </div>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue';
import { useRouter } from 'vue-router';

export default defineComponent({
  name: 'LoginView',
  setup() {
    const username = ref('');
    const password = ref('');
    const error = ref('');
    const router = useRouter();

    const login = async () => {
      try {
        const response = await fetch(`http://localhost:8888/api/user/login`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            username: username.value,
            password: password.value,
          }),
        });
        const data = await response.json();
        if (response.ok) {
          localStorage.setItem('token', data.token);
          router.push('/main');
        } else {
          error.value = '登录失败：' + data.message;
        }
      } catch (err) {
        error.value = '登录出错：' + (err as Error).message;
      }
    };

    return { username, password, error, login };
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