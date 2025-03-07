<template>
  <HomePage />
</template>

<script>
import HomePage from './components/HomePage.vue'

export default {
  components: {
    HomePage,
  },
}
</script>

<style>
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

body.dark-theme {
  background-color: #222; 
  color: #eee; 
}
</style>
