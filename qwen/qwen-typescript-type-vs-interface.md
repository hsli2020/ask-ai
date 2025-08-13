# typescript中type和interface有什么区别？

在 TypeScript 中，`type`（类型别名）和 `interface`（接口）都可以用来定义对象的结构，但它们有一些重要的区别：

## 1. 基本语法

```typescript
// interface
interface User {
  name: string;
  age: number;
}

// type
type User = {
  name: string;
  age: number;
};
```

## 2. 扩展方式不同

### Interface - 使用 `extends`
```typescript
interface Animal {
  name: string;
}

interface Dog extends Animal {
  breed: string;
}
```

### Type - 使用交叉类型 `&`
```typescript
type Animal = {
  name: string;
};

type Dog = Animal & {
  breed: string;
};
```

## 3. 声明合并（Declaration Merging）

**Interface 支持声明合并**：
```typescript
interface User {
  name: string;
}

interface User {
  age: number;
}

// 等价于
interface User {
  name: string;
  age: number;
}
```

**Type 不支持声明合并**：
```typescript
type User = {
  name: string;
};

// ❌ 错误：重复标识符 'User'
type User = {
  age: number;
};
```

## 4. 表达能力不同

**Type 更强大，可以表示更多类型**：

```typescript
// 基本类型别名
type Name = string;

// 联合类型
type Status = 'success' | 'error' | 'loading';

// 元组
type Point = [number, number];

// 映射类型
type PartialUser = Partial<User>;

// 条件类型
type IsString<T> = T extends string ? true : false;

// 接口无法直接做到这些
```

## 5. 计算属性

**Type 支持计算属性名**：
```typescript
type Keys = 'name' | 'age';
type User = {
  [K in Keys]: string;
};
```

**Interface 不支持**（除非使用索引签名）：
```typescript
interface User {
  [key: string]: string; // 只能这样，不够精确
}
```

## 6. 实现（implements）

两者都可以被类实现：
```typescript
interface Flyable {
  fly(): void;
}

type Swimmable = {
  swim(): void;
};

class Duck implements Flyable, Swimmable {
  fly() {}
  swim() {}
}
```

## 7. 性能和编译

- 两者在编译后都会被完全擦除，运行时无区别
- 在大型项目中，interface 的声明合并可能带来更好的开发体验

## 使用建议

### 优先使用 `interface` 的情况：
- 定义对象形状（特别是公共 API）
- 需要声明合并
- 面向对象编程，需要被类实现

### 使用 `type` 的情况：
- 需要联合类型、元组、基本类型别名
- 使用映射类型、条件类型等高级类型操作
- 定义函数类型

### 示例对比：

```typescript
// 推荐：对象形状用 interface
interface Person {
  name: string;
  age: number;
}

// 推荐：联合类型用 type
type ID = string | number;

// 推荐：函数类型用 type
type EventHandler = (event: Event) => void;

// 推荐：复杂类型操作用 type
type Nullable<T> = T | null;
```

总的来说，**当两者都能满足需求时，优先选择 `interface`**，因为它更符合面向对象的设计理念，且支持声明合并。当需要更复杂的类型操作时，使用 `type`。
