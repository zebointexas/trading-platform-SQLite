<template>
  <div class="container">
    <h2>Welcome!</h2>
    <button class="logout-btn" @click="logout">Log Out</button>

    <div class="card">
      <h3>Market Price</h3>
      <input v-model="pair" placeholder="Trading Pair (e.g., BTCUSD)" />
      <button @click="getPrice">Get Pricesss</button>
      <p>{{ priceResult }}</p>
    </div>

    <div class="card">
      <h3>Wallet Balance</h3>
      <input v-model="currency" placeholder="Currency (e.g., BTC)" />
      <button @click="getBalance">Get Balance</button>
      <p>{{ balanceResult }}</p>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue';
import { useRouter } from 'vue-router';

// 定义后端返回的数据结构
interface PriceResponse {
  code: number;
  msg: string;
  data?: {
    price: number;
  };
}

export default defineComponent({
  name: 'MainView',
  setup() {
    const pair = ref('');
    const currency = ref('');
    const priceResult = ref('');
    const balanceResult = ref('');
    const router = useRouter();

    const getPrice = async () => {
      try {
        const response = await fetch(
          `http://localhost:8888/api/market/price/${pair.value}`,
          {
            method: 'GET'
          }
        );
        const data: PriceResponse = await response.json();

        // 打印和显示响应数据，便于调试
        console.log('Response data:', data);
        alert(JSON.stringify(data));

        // 检查 HTTP 状态码
        if (!response.ok) {
          priceResult.value = data.msg
            ? `Failed to get price: ${data.msg}`
            : 'Failed to get price: Unknown error';
          return;
        }

        // 检查后端返回的状态码和消息
        if (data.code !== 0 || data.msg !== 'success') {
          priceResult.value = `Error: ${data.msg || 'Unknown error'}`;
          return;
        }

        // 访问嵌套的 price 字段
        if (data.data && typeof data.data.price === 'number') {
          priceResult.value = `Price: ${data.data.price}`;
        } else {
          priceResult.value = 'Error: Price not found in response';
        }
      } catch (err) {
        priceResult.value = `Error: ${(err as Error).message}`;
      }
    };

    const getBalance = async () => {
      const token = localStorage.getItem('token');

      

      if (!token) {
        balanceResult.value = 'Error: No token found. Please log in again.';
        return;
      }

      try {
        const response = await fetch(
          `http://localhost:8888/api/wallet/balance/${currency.value}`,
          {
            method: 'GET',
            headers: {
              Authorization: `Bearer ${token}`
            }
          }
        );
        const data = await response.json();

        console.log("data:", data); 

        alert(JSON.stringify(data));
        
        if (response.ok) {
          balanceResult.value = `Balance: ${data.balance} ${currency.value}`;
        } else {
          balanceResult.value = 'Failed to get balance: ' + data.message;
        }
      } catch (err) {
        balanceResult.value = 'Error: ' + (err as Error).message;
      }
    };

    const logout = () => {
      localStorage.removeItem('token');
      router.push('/');
    };

    return { pair, currency, priceResult, balanceResult, getPrice, getBalance, logout };
  },
});
</script>

<style scoped>
.container {
  max-width: 700px;
  margin: 50px auto;
  padding: 30px;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  border-radius: 15px;
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.1);
  text-align: center;
}

h2 {
  color: #2c3e50;
  font-size: 2rem;
  margin-bottom: 20px;
}

.logout-btn {
  padding: 8px 20px;
  background-color: #e74c3c;
  color: white;
  border: none;
  border-radius: 25px;
  cursor: pointer;
  transition: background-color 0.3s ease, transform 0.2s ease;
}

.logout-btn:hover {
  background-color: #c0392b;
  transform: translateY(-2px);
}

.card {
  background: white;
  padding: 20px;
  border-radius: 10px;
  margin: 20px 0;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
  transition: transform 0.3s ease;
}

.card:hover {
  transform: translateY(-5px);
}

h3 {
  color: #34495e;
  font-size: 1.5rem;
  margin-bottom: 15px;
}

input {
  display: block;
  width: 100%;
  padding: 12px;
  margin: 10px 0;
  border: 2px solid #ddd;
  border-radius: 8px;
  font-size: 1rem;
  transition: border-color 0.3s ease;
}

input:focus {
  border-color: #3498db;
  outline: none;
}

button {
  padding: 12px 25px;
  margin: 10px;
  background-color: #3498db;
  color: white;
  border: none;
  border-radius: 25px;
  font-size: 1rem;
  cursor: pointer;
  transition: background-color 0.3s ease, transform 0.2s ease;
}

button:hover {
  background-color: #2980b9;
  transform: translateY(-2px);
}

button:active {
  transform: translateY(1px);
}

p {
  margin-top: 15px;
  color: #7f8c8d;
  font-size: 1.1rem;
}
</style>