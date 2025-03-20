<template>
  <div class="home">
    <h1>Welcome to gFFL</h1>
    <p class="subtitle">GraphQL Fantasy Football League</p>
    <div v-if="loading">Loading...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <div v-else>
      <p>{{ data?.hello }}</p>
      <form @submit.prevent="handleSubmit">
        <input v-model="message" type="text" placeholder="Enter a message">
        <button type="submit">Echo</button>
      </form>
      <p v-if="echoResult">Echo result: {{ echoResult }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useQuery, useMutation } from '@vue/apollo-composable'
import gql from 'graphql-tag'

const message = ref('')
const echoResult = ref('')

const { result, loading, error } = useQuery(gql`
  query Hello {
    hello
  }
`)

const { mutate: echo } = useMutation(gql`
  mutation Echo($message: String!) {
    echo(message: $message)
  }
`)

const handleSubmit = async () => {
  try {
    const result = await echo({ message: message.value })
    echoResult.value = result.data.echo
  } catch (e) {
    console.error('Error:', e)
  }
}
</script>

<style scoped>
.home {
  text-align: center;
}

.subtitle {
  color: #666;
  margin-bottom: 2rem;
}

form {
  margin: 2rem 0;
}

input {
  padding: 0.5rem;
  margin-right: 0.5rem;
  border: 1px solid #ddd;
  border-radius: 4px;
}

button {
  padding: 0.5rem 1rem;
  background-color: #42b983;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

button:hover {
  background-color: #3aa876;
}
</style> 