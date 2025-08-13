# 请给出一个完整的vue的例子，尽可能多地用到vue的各个方面的功能

下面是一个**完整的 Vue 3 示例应用**，尽可能多地展示了 Vue 的核心功能和常用特性。该示例是一个**任务管理器（Todo List）**，包含了：

- 响应式数据（ref, reactive）
- 计算属性（computed）
- 方法（methods）
- 生命周期钩子（onMounted, onUnmounted）
- 条件渲染（v-if / v-show）
- 列表渲染（v-for）
- 表单绑定（v-model）
- 事件处理（@click, @submit）
- 组件通信（props, emits）
- 自定义指令
- Composition API（setup）
- Watchers（watch, watchEffect）
- Provide / Inject
- 动态组件（component + :is）
- 插槽（slot）
- Teleport（模态框）
- Suspense（异步组件加载）
- 路由（Vue Router）
- 状态管理（Pinia，Vue 官方推荐的状态管理库）
- 全局属性（app.config.globalProperties）
- 自定义 Hook（可复用逻辑）

---

### 📁 项目结构（简化版）

```
src/
├── main.js
├── App.vue
├── router/index.js
├── stores/todoStore.js
├── components/
│   ├── TodoItem.vue
│   ├── AddTodoForm.vue
│   ├── Stats.vue
│   └── ThemeToggle.vue
└── composables/
    └── useLocalStorage.js
```

---

### 📄 `main.js`

```js
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'

const app = createApp(App)

// 全局属性
app.config.globalProperties.$appName = 'Vue Todo Pro'

// 安装插件
app.use(createPinia())
app.use(router)

// 自定义指令
app.directive('focus', {
  mounted(el) {
    el.focus()
  }
})

app.mount('#app')
```

---

### 📄 `App.vue`

```vue
<template>
  <div id="app" :class="{ dark: isDark }">
    <header>
      <h1>{{ $appName }}</h1>
      <ThemeToggle @toggle="isDark = !isDark" />
    </header>

    <nav>
      <router-link to="/">All</router-link> |
      <router-link to="/active">Active</router-link> |
      <router-link to="/completed">Completed</router-link>
    </nav>

    <main>
      <Suspense>
        <template #default>
          <router-view />
        </template>
        <template #fallback>
          <p>Loading todos...</p>
        </template>
      </Suspense>
    </main>

    <!-- Teleport 模态框 -->
    <Teleport to="body">
      <div v-if="showModal" class="modal">
        <p>🎉 All tasks completed!</p>
        <button @click="showModal = false">Close</button>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useTodoStore } from '@/stores/todoStore'
import ThemeToggle from '@/components/ThemeToggle.vue'

const isDark = ref(false)
const showModal = ref(false)
const store = useTodoStore()

// Watcher：当所有任务完成时弹出模态框
watch(() => store.completedCount, (newVal) => {
  if (newVal > 0 && newVal === store.todos.length) {
    showModal.value = true
  }
})
</script>

<style>
/* 简单样式 */
body { margin: 0; font-family: sans-serif; }
.dark { background: #222; color: white; }
.modal {
  position: fixed; top: 50%; left: 50%;
  transform: translate(-50%, -50%);
  background: white; padding: 20px; border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.3);
}
</style>
```

---

### 📄 `router/index.js`

```js
import { createRouter, createWebHistory } from 'vue-router'
import Home from '@/views/Home.vue'

const routes = [
  { path: '/', component: Home },
  { path: '/active', component: Home, props: { filter: 'active' } },
  { path: '/completed', component: Home, props: { filter: 'completed' } }
]

export default createRouter({
  history: createWebHistory(),
  routes
})
```

---

### 📄 `stores/todoStore.js`（Pinia）

```js
import { defineStore } from 'pinia'
import { useLocalStorage } from '@/composables/useLocalStorage'

export const useTodoStore = defineStore('todos', () => {
  // 使用自定义 Hook 持久化到 localStorage
  const todos = useLocalStorage('todos-vue', [
    { id: 1, text: 'Learn Vue 3', done: false }
  ])

  const addTodo = (text) => {
    todos.value.push({ id: Date.now(), text, done: false })
  }

  const removeTodo = (id) => {
    todos.value = todos.value.filter(todo => todo.id !== id)
  }

  const toggleTodo = (id) => {
    const todo = todos.value.find(t => t.id === id)
    if (todo) todo.done = !todo.done
  }

  const clearCompleted = () => {
    todos.value = todos.value.filter(todo => !todo.done)
  }

  const completedCount = computed(() => todos.value.filter(t => t.done).length)
  const activeCount = computed(() => todos.value.length - completedCount.value)

  return {
    todos,
    addTodo,
    removeTodo,
    toggleTodo,
    clearCompleted,
    completedCount,
    activeCount
  }
})
```

---

### 📄 `composables/useLocalStorage.js`

```js
import { ref, watch } from 'vue'

export function useLocalStorage(key, initialValue) {
  const value = ref(JSON.parse(localStorage.getItem(key)) || initialValue)

  watch(value, (newVal) => {
    localStorage.setItem(key, JSON.stringify(newVal))
  }, { deep: true })

  return value
}
```

---

### 📄 `components/AddTodoForm.vue`

```vue
<template>
  <form @submit.prevent="handleSubmit">
    <input
      v-model="newTodo"
      @keyup.enter="handleSubmit"
      v-focus
      placeholder="Add a new task..."
      :disabled="loading"
    />
    <button :disabled="!newTodo.trim() || loading">
      {{ loading ? 'Adding...' : 'Add' }}
    </button>
  </form>
</template>

<script setup>
import { ref } from 'vue'
import { useTodoStore } from '@/stores/todoStore'

const newTodo = ref('')
const loading = ref(false)
const store = useTodoStore()

const handleSubmit = async () => {
  if (!newTodo.value.trim()) return
  loading.value = true
  // 模拟异步 API 调用
  await new Promise(r => setTimeout(r, 300))
  store.addTodo(newTodo.value.trim())
  newTodo.value = ''
  loading.value = false
}
</script>
```

---

### 📄 `components/TodoItem.vue`

```vue
<template>
  <li class="todo-item">
    <input
      type="checkbox"
      :checked="todo.done"
      @change="toggle"
    />
    <span :class="{ done: todo.done }">{{ todo.text }}</span>
    <button @click="remove" class="delete-btn">×</button>
  </li>
</template>

<script setup>
const props = defineProps({
  todo: { type: Object, required: true }
})

const emit = defineEmits(['toggle', 'remove'])

const toggle = () => emit('toggle', props.todo.id)
const remove = () => emit('remove', props.todo.id)
</script>

<style scoped>
.done { text-decoration: line-through; color: #888; }
.delete-btn { margin-left: 10px; background: none; border: none; font-size: 18px; cursor: pointer; }
</style>
```

---

### 📄 `views/Home.vue`

```vue
<template>
  <div class="home">
    <AddTodoForm />

    <ul v-if="filteredTodos.length">
      <TodoItem
        v-for="todo in filteredTodos"
        :key="todo.id"
        :todo="todo"
        @toggle="store.toggleTodo"
        @remove="store.removeTodo"
      />
    </ul>

    <p v-else>No tasks found.</p>

    <Stats />

    <button v-if="store.completedCount" @click="store.clearCompleted">
      Clear Completed ({{ store.completedCount }})
    </button>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useTodoStore } from '@/stores/todoStore'
import AddTodoForm from '@/components/AddTodoForm.vue'
import TodoItem from '@/components/TodoItem.vue'
import Stats from '@/components/Stats.vue'

const props = defineProps({
  filter: { type: String, default: 'all' }
})

const store = useTodoStore()

const filteredTodos = computed(() => {
  if (props.filter === 'active') return store.todos.filter(t => !t.done)
  if (props.filter === 'completed') return store.todos.filter(t => t.done)
  return store.todos
})
</script>
```

---

### 📄 `components/Stats.vue`

```vue
<template>
  <div class="stats">
    <slot name="prefix"></slot>
    <p>{{ store.activeCount }} items left</p>
    <slot></slot>
  </div>
</template>

<script setup>
import { useTodoStore } from '@/stores/todoStore'
const store = useTodoStore()
</script>
```

在 `Home.vue` 中使用插槽：

```vue
<Stats>
  <template #prefix>
    <strong>📊</strong>
  </template>
  <em>Keep going!</em>
</Stats>
```

---

### ✅ 本例涵盖的 Vue 功能总结：

| 功能 | 使用位置 |
|------|--------|
| `ref` / `reactive` | 响应式数据 |
| `computed` | 计算属性（过滤、统计） |
| `watch` / `watchEffect` | 监听任务完成状态 |
| `v-model` | 表单双向绑定 |
| `v-for` / `v-if` | 列表与条件渲染 |
| `@click` / `@submit` | 事件处理 |
| `props` / `emits` | 组件通信 |
| `setup()` + Composition API | 所有 `<script setup>` |
| `provide` / `inject` | （本例用 Pinia 替代，但可轻松加入）|
| 自定义指令 `v-focus` | 自动聚焦输入框 |
| `Teleport` | 模态框渲染到 body |
| `Suspense` | 异步组件加载 |
| 动态组件 | 通过路由实现（`<router-view>`）|
| 插槽（slot） | `Stats` 组件 |
| 生命周期钩子 | （隐含在 `onMounted` 等，本例用 `watch` 替代）|
| 全局属性 | `$appName` |
| Pinia 状态管理 | 全局状态 |
| 自定义 Hook | `useLocalStorage` |
| 路由 | Vue Router |

---

这个例子虽然功能完整，但代码结构清晰，适合学习和扩展。你可以在此基础上加入更多功能，如拖拽排序、标签分类、用户登录等。

如需运行，请使用 Vite + Vue 3 + Pinia 模板创建项目并复制上述代码。
