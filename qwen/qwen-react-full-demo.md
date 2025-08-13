# 请给出一个完整的react的例子，尽可能多地展示react的各个方面的功能

以下是一个完整的 React 示例，展示了 React 的核心功能和最佳实践：

```jsx
import React, { useState, useEffect, useContext, useReducer, useCallback, useMemo, useRef, createContext } from 'react';
import ReactDOM from 'react-dom/client';

// 1. Context API - 创建上下文
const ThemeContext = createContext();
const UserContext = createContext();

// 2. 自定义 Hook
const useLocalStorage = (key, initialValue) => {
  const [storedValue, setStoredValue] = useState(() => {
    try {
      const item = window.localStorage.getItem(key);
      return item ? JSON.parse(item) : initialValue;
    } catch (error) {
      console.log(error);
      return initialValue;
    }
  });

  const setValue = (value) => {
    try {
      setStoredValue(value);
      window.localStorage.setItem(key, JSON.stringify(value));
    } catch (error) {
      console.log(error);
    }
  };

  return [storedValue, setValue];
};

// 3. Reducer Hook - 复杂状态管理
const todoReducer = (state, action) => {
  switch (action.type) {
    case 'ADD_TODO':
      return [...state, { id: Date.now(), text: action.payload, completed: false }];
    case 'TOGGLE_TODO':
      return state.map(todo =>
        todo.id === action.payload ? { ...todo, completed: !todo.completed } : todo
      );
    case 'DELETE_TODO':
      return state.filter(todo => todo.id !== action.payload);
    case 'CLEAR_COMPLETED':
      return state.filter(todo => !todo.completed);
    default:
      return state;
  }
};

// 4. 子组件 - 展示 Props、事件处理、条件渲染
const TodoItem = ({ todo, onToggle, onDelete }) => {
  // 5. useRef - 获取 DOM 引用
  const todoRef = useRef(null);

  // 6. useEffect - 副作用处理
  useEffect(() => {
    if (todo.completed) {
      todoRef.current.style.textDecoration = 'line-through';
    } else {
      todoRef.current.style.textDecoration = 'none';
    }
  }, [todo.completed]);

  return (
    <div className="todo-item">
      <input
        type="checkbox"
        checked={todo.completed}
        onChange={() => onToggle(todo.id)}
      />
      <span ref={todoRef}>{todo.text}</span>
      <button onClick={() => onDelete(todo.id)}>删除</button>
    </div>
  );
};

// 7. React.memo - 性能优化
const TodoList = React.memo(({ todos, onToggle, onDelete }) => {
  return (
    <div className="todo-list">
      {todos.map(todo => (
        <TodoItem
          key={todo.id}
          todo={todo}
          onToggle={onToggle}
          onDelete={onDelete}
        />
      ))}
    </div>
  );
});

// 8. 主应用组件
const App = () => {
  // 9. useState - 基础状态管理
  const [newTodo, setNewTodo] = useState('');
  const [theme, setTheme] = useState('light');
  const [user, setUser] = useLocalStorage('user', { name: '', isLoggedIn: false });

  // 10. useReducer - 复杂状态管理
  const [todos, dispatch] = useReducer(todoReducer, []);

  // 11. useCallback - 优化函数引用
  const handleAddTodo = useCallback(() => {
    if (newTodo.trim()) {
      dispatch({ type: 'ADD_TODO', payload: newTodo.trim() });
      setNewTodo('');
    }
  }, [newTodo]);

  const handleToggleTodo = useCallback((id) => {
    dispatch({ type: 'TOGGLE_TODO', payload: id });
  }, []);

  const handleDeleteTodo = useCallback((id) => {
    dispatch({ type: 'DELETE_TODO', payload: id });
  }, []);

  const handleClearCompleted = useCallback(() => {
    dispatch({ type: 'CLEAR_COMPLETED' });
  }, []);

  // 12. useMemo - 优化计算值
  const completedCount = useMemo(() => {
    return todos.filter(todo => todo.completed).length;
  }, [todos]);

  const activeCount = useMemo(() => {
    return todos.length - completedCount;
  }, [todos, completedCount]);

  // 13. useEffect - 生命周期和副作用
  useEffect(() => {
    document.title = `待办事项 (${activeCount} 项待完成)`;
  }, [activeCount]);

  // 14. 条件渲染
  const renderLogin = () => {
    if (user.isLoggedIn) {
      return (
        <div>
          <p>欢迎，{user.name}！</p>
          <button onClick={() => setUser({ name: '', isLoggedIn: false })}>
            退出登录
          </button>
        </div>
      );
    } else {
      return (
        <div>
          <input
            type="text"
            placeholder="输入用户名"
            value={user.name}
            onChange={(e) => setUser({ ...user, name: e.target.value })}
          />
          <button onClick={() => setUser({ ...user, isLoggedIn: true })}>
            登录
          </button>
        </div>
      );
    }
  };

  // 15. 事件处理
  const handleKeyPress = (e) => {
    if (e.key === 'Enter') {
      handleAddTodo();
    }
  };

  // 16. 列表渲染
  const renderStats = () => (
    <div className="stats">
      <span>总计: {todos.length}</span>
      <span>完成: {completedCount}</span>
      <span>待办: {activeCount}</span>
    </div>
  );

  return (
    // 17. Context Provider - 提供上下文值
    <ThemeContext.Provider value={{ theme, toggleTheme: () => setTheme(theme === 'light' ? 'dark' : 'light') }}>
      <UserContext.Provider value={{ user, setUser }}>
        <div className={`app ${theme}`}>
          <header>
            <h1>React 功能演示应用</h1>
            <button onClick={() => setTheme(theme === 'light' ? 'dark' : 'light')}>
              切换主题: {theme}
            </button>
          </header>

          <main>
            {/* 18. 表单处理 */}
            <div className="input-section">
              <input
                type="text"
                value={newTodo}
                onChange={(e) => setNewTodo(e.target.value)}
                onKeyPress={handleKeyPress}
                placeholder="添加新的待办事项..."
              />
              <button onClick={handleAddTodo}>添加</button>
            </div>

            {/* 19. 条件渲染 - 显示/隐藏 */}
            {todos.length > 0 && (
              <>
                <TodoList
                  todos={todos}
                  onToggle={handleToggleTodo}
                  onDelete={handleDeleteTodo}
                />
                {renderStats()}
                {completedCount > 0 && (
                  <button onClick={handleClearCompleted} className="clear-btn">
                    清除已完成 ({completedCount})
                  </button>
                )}
              </>
            )}

            {/* 20. 用户认证状态 */}
            <div className="user-section">
              <h3>用户管理</h3>
              {renderLogin()}
            </div>

            {/* 21. 错误边界（简化版） */}
            {user.name && user.name.length > 20 && (
              <div className="error">
                用户名不能超过20个字符！
              </div>
            )}
          </main>
        </div>
      </UserContext.Provider>
    </ThemeContext.Provider>
  );
};

// 22. 自定义 Hook 使用示例
const ThemeToggle = () => {
  const { theme, toggleTheme } = useContext(ThemeContext);
  return (
    <button onClick={toggleTheme}>
      当前主题: {theme}
    </button>
  );
};

// 23. 错误边界组件（实际项目中应该这样实现）
class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error) {
    return { hasError: true };
  }

  componentDidCatch(error, errorInfo) {
    console.log('Error caught:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return <h2>出错了！</h2>;
    }
    return this.props.children;
  }
}

// 24. 最终渲染
const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  <ErrorBoundary>
    <App />
  </ErrorBoundary>
);

// 25. CSS 样式（内联样式示例）
const styles = {
  app: {
    padding: '20px',
    fontFamily: 'Arial, sans-serif'
  },
  light: {
    backgroundColor: '#fff',
    color: '#333'
  },
  dark: {
    backgroundColor: '#333',
    color: '#fff'
  }
};
```

## 这个示例展示了 React 的以下核心功能：

### **状态管理**
- `useState` - 基础状态管理
- `useReducer` - 复杂状态逻辑
- `useContext` - 全局状态共享

### **副作用处理**
- `useEffect` - 生命周期和副作用
- 清理函数（虽然示例中没直接展示，但可以添加）

### **性能优化**
- `useCallback` - 防止不必要的函数重新创建
- `useMemo` - 缓存计算结果
- `React.memo` - 组件记忆化

### **引用和 DOM 操作**
- `useRef` - 访问 DOM 元素和保持可变值

### **自定义 Hook**
- `useLocalStorage` - 封装 localStorage 逻辑

### **组件模式**
- 函数组件
- 条件渲染
- 列表渲染
- 表单处理
- 事件处理

### **高级特性**
- Context API - 跨组件通信
- 错误边界 - 错误处理
- Portals（虽然没展示，但可以添加）

### **最佳实践**
- 组件拆分
- 状态提升
- 单向数据流
- 不可变数据更新

这个示例涵盖了 React 开发中的大部分常用功能，可以作为学习和参考的完整模板。
