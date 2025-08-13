# 总结一下在typescript中定义类型的所有方式，并给出用法示例

在 TypeScript 中，类型系统是构建健壮、可维护代码的核心。以下是 TypeScript 中定义类型的所有主要方式，并附带用法示例，帮助你全面掌握类型定义技巧。

---

## 1. 基本类型

```typescript
let isDone: boolean = false;
let count: number = 42;
let name: string = "Alice";
let list: number[] = [1, 2, 3];
let tuple: [string, number] = ["hello", 10];
let anyValue: any = "anything";
let unknownValue: unknown = "safe";
let nothing: void = undefined;
let neverValue: never = (() => { throw new Error(); })();
```

---

## 2. 接口（interface）

```typescript
interface User {
  id: number;
  name: string;
  email?: string; // 可选属性
  readonly createdAt: Date; // 只读属性
}

interface Point {
  x: number;
  y: number;
}

interface Draggable {
  drag(): void;
}

interface Movable {
  move(): void;
}

interface Widget extends Draggable, Movable {
  render(): void;
}
```

---

## 3. 类型别名（type）

```typescript
type ID = string | number;
type Point = { x: number; y: number };
type Callback = (data: string) => void;
type Role = "admin" | "user" | "guest";
type DeepPartial<T> = {
  [P in keyof T]?: DeepPartial<T[P]>;
};
```

---

## 4. 枚举（enum）

```typescript
enum Direction {
  Up = "UP",
  Down = "DOWN",
  Left = "LEFT",
  Right = "RIGHT"
}

enum Status {
  Pending,
  Success,
  Error
}

const enum Color {
  Red = 1,
  Green,
  Blue
}
```

---

## 5. 联合类型（Union Types）

```typescript
type ID = string | number;
type Shape = Circle | Square;

function printId(id: ID) {
  console.log(id);
}

function move(arg: string | number) {
  if (typeof arg === "string") {
    // string 逻辑
  } else {
    // number 逻辑
  }
}
```

---

## 6. 交叉类型（Intersection Types）

```typescript
type A = { a: string };
type B = { b: number };
type C = A & B; // { a: string; b: number }

interface Clickable {
  onClick: () => void;
}

interface Scrollable {
  onScroll: () => void;
}

type UIElement = Clickable & Scrollable;
```

---

## 7. 元组（Tuple）

```typescript
let tuple: [string, number] = ["hello", 42];
let readonlyTuple: readonly [number, string] = [1, "a"];

type Person = [name: string, age: number];
```

---

## 8. 数组类型

```typescript
let numbers: number[] = [1, 2, 3];
let strings: Array<string> = ["a", "b"];
let readonlyArray: readonly number[] = [1, 2, 3];
```

---

## 9. 函数类型

```typescript
type Add = (a: number, b: number) => number;
const add: Add = (a, b) => a + b;

function greet(name: string, age?: number): string {
  return `Hello ${name}`;
}

function log(msg: string): void {
  console.log(msg);
}
```

---

## 10. 泛型（Generics）

```typescript
function identity<T>(arg: T): T {
  return arg;
}

interface Box<T> {
  value: T;
}

type Pair<T, U> = [T, U];

function merge<T, U>(a: T, b: U): T & U {
  return { ...a, ...b };
}
```

---

## 11. 类型推断

```typescript
let x = 10; // number
let y = "hello"; // string
let z = [1, 2, 3]; // number[]

function add(a: number, b: number) {
  return a + b; // 返回类型自动推断为 number
}
```

---

## 12. 类型断言

```typescript
let someValue: any = "hello";
let strLength: number = (someValue as string).length;
let len = (<string>someValue).length;

const canvas = document.getElementById("canvas") as HTMLCanvasElement;
```

---

## 13. 工具类型（Utility Types）

```typescript
type PartialUser = Partial<User>;
type ReadonlyUser = Readonly<User>;
type RequiredUser = Required<User>;
type PickUser = Pick<User, "id" | "name">;
type OmitUser = Omit<User, "email">;
type RecordType = Record<string, number>;
type NonNullableType = NonNullable<string | null>;
type ReturnTypeFunc = ReturnType<() => string>;
type ParametersFunc = Parameters<(a: number, b: string) => void>;
```

---

## 14. 映射类型（Mapped Types）

```typescript
type Readonly<T> = {
  readonly [P in keyof T]: T[P];
};

type Nullable<T> = {
  [P in keyof T]: T[P] | null;
};

type Keys = "option1" | "option2";
type Flags = { [K in Keys]: boolean };
```

---

## 15. 条件类型（Conditional Types）

```typescript
type IsString<T> = T extends string ? true : false;

type ExtractString<T> = T extends string ? T : never;

type ReturnType<T> = T extends (...args: any[]) => infer R ? R : any;

type Flatten<T> = T extends Array<infer U> ? U : T;
```

---

## 16. 字面量类型（Literal Types）

```typescript
type Direction = "up" | "down" | "left" | "right";
type Status = 200 | 404 | 500;

function move(direction: Direction) {}
```

---

## 17. 索引类型（Index Types）

```typescript
interface Person {
  name: string;
  age: number;
}

type PersonKeys = keyof Person; // "name" | "age"
type PersonValues = Person[keyof Person]; // string | number

function getValue<T, K extends keyof T>(obj: T, key: K): T[K] {
  return obj[key];
}
```

---

## 18. 命名空间（Namespace）

```typescript
namespace MathUtils {
  export const PI = 3.14159;
  export function add(a: number, b: number): number {
    return a + b;
  }
}

MathUtils.add(1, 2);
```

---

## 19. 类型守卫（Type Guards）

```typescript
function isString(value: any): value is string {
  return typeof value === "string";
}

function process(value: string | number) {
  if (isString(value)) {
    console.log(value.toUpperCase());
  } else {
    console.log(value.toFixed(2));
  }
}

// instanceof 守卫
class Dog { bark() {} }
class Cat { meow() {} }

function makeSound(animal: Dog | Cat) {
  if (animal instanceof Dog) {
    animal.bark();
  } else {
    animal.meow();
  }
}

// in 守卫
interface A { a: string }
interface B { b: number }

function print(obj: A | B) {
  if ("a" in obj) {
    console.log(obj.a);
  } else {
    console.log(obj.b);
  }
}
```

---

## 20. 声明合并与模块类型

```typescript
// 接口合并
interface Box {
  height: number;
}
interface Box {
  width: number;
}
// 合并后: { height: number; width: number }

// 模块声明
declare module "my-module" {
  export function hello(): void;
}

// 全局变量声明
declare const VERSION: string;
```

---

## 21. 高级类型技巧

```typescript
// 模板字面量类型
type EventName = `on${Capitalize<string>}`;

// 递归类型
type Json =
  | string
  | number
  | boolean
  | null
  | Json[]
  | { [key: string]: Json };

// 排除类型
type Filtered = Exclude<"a" | "b" | "c", "a">; // "b" | "c"

// 条件映射
type Mutable<T> = {
  -readonly [P in keyof T]: T[P];
};

type Optional<T> = {
  [P in keyof T]?: T[P];
};
```

---

## 最佳实践建议

- 优先使用 `interface` 定义对象结构，支持扩展和合并。
- 使用 `type` 表示联合、交叉、映射等复杂类型。
- 避免滥用 `any`，优先使用 `unknown`。
- 使用泛型提升复用性。
- 使用工具类型简化类型操作。
- 使用类型守卫提升类型安全。
- 使用 `as const` 保留字面量类型。

---

通过掌握以上类型定义方式，你可以构建出类型安全、可维护且高度可复用的 TypeScript 代码。
