<template>
  <div class="container">
    <h2>登录</h2>
    <input v-model="username" placeholder="用户名" />
    <input v-model="password" type="password" placeholder="密码" />
    <button @click="login">登录</button>
    <p v-if="error">{{ error }}</p>
    <p>没有账号？<router-link to="/register">去注册</router-link></p>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue';
import { useRouter } from 'vue-router';

// 定义后端返回的数据结构
interface LoginResponse {
  code: number;
  msg: string;
  data?: {
    token: string;
  };
}

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
        const data: LoginResponse = await response.json();

        // 打印 data 的完整结构
        console.log('Login response data:', data);
        console.log('Data type:', typeof data);
        console.log('Response status:', response.status, 'OK:', response.ok);

        // 显示格式化的 JSON 字符串
        // alert(JSON.stringify(data, null, 2));

        // 检查 HTTP 状态码
        if (!response.ok) {
          error.value = data.msg
            ? `登录失败：${data.msg}`
            : '登录失败：未知错误';
          return;
        }

        // 检查后端返回的状态码和消息
        if (data.code !== 0 || data.msg !== 'success') {
          error.value = `登录失败：${data.msg || '未知错误'}`;
          return;
        }

        // 访问嵌套的 token 字段
        if (data.data && typeof data.data.token === 'string') {
          localStorage.setItem('token', data.data.token);
          router.push('/main');
        } else {
          error.value = '登录失败：未找到 token';
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