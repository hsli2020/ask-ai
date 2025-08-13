<template>
  <div>
    <button @click="increaseCounter" class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded">Increase Counter</button>
    <p class="text-gray-700 text-base">Counter: {{ counter }}</p>
    
    <input type="text" v-model="inputText" class="border-2 border-gray-300 p-2 rounded" />
    <p class="text-gray-700 text-base">Input: {{ inputText }}</p>
  </div>
</template>

<script setup>
import { ref } from 'vue';

const counter = ref(0);
const inputText = ref('');

const increaseCounter = () => {
  counter++;
};
</script>
