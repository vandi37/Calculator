<template>
  <div class="home-page">
    <button id="theme-toggle" @click="toggleTheme">Switch theme</button>

    <div class="container">
      <h1>Cute calculator</h1>
      <input
        type="text"
        id="expression"
        placeholder="Enter your expression"
        v-model="expression"
        @keyup.enter="submitExpression"
      />
      <button @click="submitExpression">Send</button>
    </div>

    <h2>Expression history</h2>
    <ul id="expressionList">
      <li v-for="(item, index) in expressionHistory" :key="index">
        <span>{{ item.id }}</span>
        <span>{{ item.expression }}</span>
        <span
          class="status"
          :class="{
            'status-nothing': item.status === 'Nothing',
            'status-processing': item.status === 'Processing',
            'status-error': item.status === 'Error',
            'status-finished': item.status === 'Finished',
            'status-waiting': item.status === 'Waiting',
          }"
          >{{ item.statusDisplay }}</span
        >
      </li>
    </ul>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  data() {
    return {
      expression: '',
      expressionHistory: [],
      isDarkTheme: window.matchMedia('(prefers-color-scheme: dark)').matches,
      operators: {
        0: '+',
        2: '-',
        8: '*',
        24: '/',
      },
      apiUrl: `http://localhost:${this.$config.port.api}${this.$config.path.calc}`,
      apiStatusUrl: `http://localhost:${this.$config.port.api}${this.$config.path.get}`,
      intervalId: null,
    }
  },
  mounted() {
    this.applyTheme()
    this.refreshStatus()
    this.intervalId = setInterval(() => {
      this.refreshStatus()
    }, 100)
  },
  beforeUnmount() {
    if (this.intervalId) {
      clearInterval(this.intervalId)
      this.intervalId = null
    }
  },
  methods: {
    async submitExpression() {
      if (this.expression.trim() === '') return

      try {
        const response = await axios.post(this.apiUrl, {
          expression: this.expression,
        })

        alert(`Expression submitted with ID: ${response.data.id}`)
        this.expression = ''
        this.refreshStatus()
      } catch (error) {
        alert(`Error: ${error.message}`)
      }
    },

    formatStatusDisplay(status) {
      if (status.value === null) {
        return status.status
      }

      if (typeof status.value === 'string' || typeof status.value === 'number') {
        return `${status.status}: ${status.value}`
      }

      if (status.value && typeof status.value === 'object') {
        const { arg1, arg2, operation } = status.value
        const operator = this.operators[operation] || '?'
        return `${status.status}: ${arg1} ${operator} ${arg2}`
      }

      return `${status.status}: unknown value`
    },

    async refreshStatus() {
      try {
        const response = await axios.get(this.apiStatusUrl)
        this.expressionHistory = response.data.expressions.reverse().map((status) => ({
          id: status.id,
          expression: status.expression,
          status: status.status,
          statusDisplay: this.formatStatusDisplay(status),
        }))
      } catch (error) {
        alert(`Error fetching status: ${error.message}`)
      }
    },
    toggleTheme() {
      this.isDarkTheme = !this.isDarkTheme
      this.applyTheme()
    },

    applyTheme() {
      const body = document.body
      body.classList.toggle('dark-theme', this.isDarkTheme)
      body.classList.toggle('light-theme', !this.isDarkTheme)
    },
  },
}
</script>

<style scoped>
body {
  font-family: sans-serif;
  background-color: #ffe6f2;
  color: #333;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-start;
  min-height: 100vh;
  margin: 0;
  padding-top: 20px;
  transition:
    background-color 0.3s ease,
    color 0.3s ease;
}

.container {
  background-color: #fff0f5;
  border-radius: 10px;
  padding: 20px;
  box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
  width: 80%;
  max-width: 600px;
  margin-bottom: 20px;
  transition: background-color 0.3s ease;
}

h1 {
  color: #ff69b4;
  text-align: center;
}

input[type='text'] {
  width: 100%;
  padding: 10px;
  margin-bottom: 10px;
  border: 1px solid #ffb6c1;
  border-radius: 5px;
  box-sizing: border-box;
  transition:
    border-color 0.3s ease,
    background-color 0.3s ease,
    color 0.3s ease;
}

button {
  background-color: #ff69b4;
  color: white;
  padding: 10px 20px;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  transition: background-color 0.3s ease;
}

button:hover {
  background-color: #ff85c0;
}

ul {
  list-style: none;
  padding: 0;
}

li {
  background-color: #ffe4e1;
  padding: 10px;
  margin-bottom: 5px;
  border-radius: 5px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  transition:
    background-color 0.3s ease,
    color 0.3s ease;
}

li span:first-child {
  font-weight: bold;
  margin-right: 10px;
}

.status {
  font-style: italic;
}

.status-nothing {
  color: #908a7f;
}

.status-finished {
  color: #28a745;
}

.status-error {
  color: #dc3545;
}

.status-processing {
  color: #17a2b8;
}

.status-waiting {
  color: #f97800;
}

body.dark-theme {
  background-color: #222;
  color: #eee;
}

body.dark-theme .container {
  background-color: #333;
  box-shadow: 0 0 10px rgba(255, 255, 255, 0.1);
}

body.dark-theme input[type='text'] {
  border-color: #555;
  background-color: #444;
  color: #eee;
}

body.dark-theme li {
  background-color: #444;
  color: #eee;
}

body.dark-theme .status-nothing {
  color: #d7d7d7;
}

body.dark-theme .status-finished {
  color: #b2ff59;
}

body.dark-theme .status-error {
  color: #f8bbd0;
}

body.dark-theme .status-processing {
  color: #80deea;
}

body.dark-theme .status-waiting {
  color: #f9e368;
}

#theme-toggle {
  background-color: #ff69b4;
  color: white;
  padding: 10px 20px;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  transition: background-color 0.3s ease;
  margin-bottom: 10px;
}

#theme-toggle:hover {
  background-color: #ff85c0;
}
</style>
