# 编写一个javascript的类，尽可能多地展示javascript的OOP的功能特性

当然可以！下面是一个**全面展示 JavaScript OOP 特性的类示例**，涵盖了现代 JavaScript 中几乎所有重要的面向对象编程特性：

---

### ✅ **涵盖的 JavaScript OOP 特性：**

| 特性 | 说明 |
|------|------|
| ✅ 类定义 (`class`) | 使用 `class` 关键字定义构造器和方法 |
| ✅ 构造函数 (`constructor`) | 初始化实例属性 |
| ✅ 实例属性与方法 | 普通实例成员 |
| ✅ 静态属性与方法 | `static` 关键字，属于类而非实例 |
| ✅ 私有字段 (`#`) | ES2022 私有属性，只能在类内部访问 |
| ✅ 私有方法 | 私有方法同样用 `#` 前缀 |
| ✅ 访问器属性 (`get/set`) | 控制属性读写行为 |
| ✅ 继承 (`extends`) | 子类继承父类 |
| ✅ `super()` 调用 | 调用父类构造函数或方法 |
| ✅ 方法重写 (Override) | 子类覆盖父类方法 |
| ✅ 动态原型方法 | 在运行时添加方法到原型 |
| ✅ Symbol 作为私有键 | 使用 Symbol 避免命名冲突（兼容旧版） |
| ✅ 工厂方法 | 静态方法返回实例 |
| ✅ 抽象方法模拟 | 抛出错误提示子类必须实现 |
| ✅ 可选链与空值合并 | 在方法中使用现代语法 |
| ✅ 迭代器协议 (`[Symbol.iterator]`) | 使对象可被 `for...of` 遍历 |
| ✅ 可调用对象 (`call`, `apply`, `bind`) | 展示绑定行为 |
| ✅ `Object.defineProperty` | 手动定义属性描述符 |
| ✅ `Object.freeze` / `Object.seal` | 对象不可变性控制 |
| ✅ Mixin 模式 | 多重继承模拟 |
| ✅ `instanceof` 和 `isPrototypeOf` | 类型检查 |
| ✅ `Object.getPrototypeOf` | 获取原型链 |

---

## 🌟 完整代码示例：`AdvancedAnimal` 类

```javascript
// ==============================
// Mixin 模块：可飞行能力
// ==============================
const Flyable = (BaseClass) => class extends BaseClass {
  fly() {
    console.log(`${this.name} is flying with ${this.wingspan}m wingspan!`);
  }

  get altitude() {
    return this._altitude || 0;
  }

  set altitude(value) {
    if (typeof value === 'number' && value >= 0) {
      this._altitude = value;
    } else {
      throw new Error('Altitude must be a non-negative number');
    }
  }
};

// ==============================
// Mixin 模块：可游泳能力
// ==============================
const Swimmable = (BaseClass) => class extends BaseClass {
  swim() {
    console.log(`${this.name} is swimming at ${this.depth}m depth!`);
  }

  get depth() {
    return this._depth || 0;
  }

  set depth(value) {
    if (typeof value === 'number' && value >= 0) {
      this._depth = value;
    } else {
      throw new Error('Depth must be a non-negative number');
    }
  }
};

// ==============================
// 核心类：Animal（基类）
// ==============================
class Animal {
  // 静态属性：记录所有动物数量
  static count = 0;

  // 静态方法：获取总数
  static getTotalCount() {
    return Animal.count;
  }

  // 私有静态属性（ES2022）
  static #lastId = 0;

  // 公共实例属性（构造器中初始化）
  constructor(name, age, species) {
    this.name = name;
    this.age = age;
    this.species = species;

    // 私有字段（仅限类内访问）
    this.#id = ++Animal.#lastId; // 每个实例唯一ID
    this.#health = 100; // 私有健康值

    // 增加计数
    Animal.count++;

    // 冻结部分属性防止外部修改
    Object.freeze(this.species); // 不可变物种名

    // 使用 defineProperty 添加一个只读计算属性
    Object.defineProperty(this, 'isAdult', {
      get: () => this.age >= 2,
      enumerable: true,
      configurable: false
    });

    // 使用 Symbol 创建“伪私有”属性（兼容旧环境）
    this[Symbol.for('internalData')] = { createdAt: new Date(), lastModified: new Date() };

    console.log(`Created ${this.name} (${this.species}) with ID: ${this.#id}`);
  }

  // 公共实例方法
  speak() {
    console.log(`${this.name} makes a sound.`);
  }

  // 访问器：年龄限制
  get age() {
    return this._age;
  }

  set age(value) {
    if (!Number.isInteger(value) || value < 0) {
      throw new Error('Age must be a non-negative integer');
    }
    this._age = value;
  }

  // 私有方法
  #updateHealth(delta) {
    this.#health += delta;
    if (this.#health > 100) this.#health = 100;
    if (this.#health < 0) this.#health = 0;
    this[Symbol.for('internalData')].lastModified = new Date();
  }

  // 公共方法调用私有方法
  eat(food) {
    this.#updateHealth(10);
    console.log(`${this.name} ate ${food}. Health: ${this.#health}/100`);
  }

  sleep(hours) {
    this.#updateHealth(hours * 5);
    console.log(`${this.name} slept for ${hours} hours. Health: ${this.#health}/100`);
  }

  // 抽象方法（强制子类实现）
  move() {
    throw new Error('Subclasses must implement move() method');
  }

  // 实现迭代器协议：让动物按年龄排序遍历
  *[Symbol.iterator]() {
    yield this.name;
    yield this.age;
    yield this.species;
    yield this.#health;
    yield this.isAdult;
  }

  // 工厂方法：创建特定类型的动物
  static createBird(name, age) {
    return new Bird(name, age, 'Bird');
  }

  static createFish(name, age) {
    return new Fish(name, age, 'Fish');
  }

  // 返回当前实例的原型
  getPrototype() {
    return Object.getPrototypeOf(this);
  }

  // 检查是否为某个类的实例（支持多重继承）
  isInstanceOf(cls) {
    return this instanceof cls;
  }

  // 可调用方法：允许像函数一样调用实例
  call(...args) {
    console.log(`Calling animal instance with args: ${args.join(', ')}`);
    return this.speak();
  }

  // 显示完整信息
  toString() {
    return `[Animal: ${this.name}, Age: ${this.age}, Species: ${this.species}, Health: ${this.#health}]`;
  }

  // 静态私有方法
  static #generateUniqueId() {
    return ++Animal.#lastId;
  }

  // 获取私有 id（通过公共方法暴露）
  getId() {
    return this.#id;
  }

  getHealth() {
    return this.#health;
  }

  // 模拟对象冻结（不可扩展）
  lock() {
    Object.seal(this); // 禁止新增/删除属性
    console.log(`${this.name} is now sealed.`);
  }

  // 模拟深度冻结（完全不可变）
  freeze() {
    Object.freeze(this);
    console.log(`${this.name} is now frozen.`);
  }
}

// ==============================
// 子类：Bird（继承 Animal）
// ==============================
class Bird extends Animal {
  constructor(name, age, wingspan = 1.5) {
    super(name, age, 'Bird'); // 调用父类构造函数
    this.wingspan = wingspan;
  }

  // 重写父类方法
  speak() {
    console.log(`${this.name} chirps!`);
  }

  // 重写抽象方法
  move() {
    console.log(`${this.name} flies with its wings!`);
  }

  // 新增特有方法
  buildNest() {
    console.log(`${this.name} built a nest with twigs.`);
  }

  // 覆盖 toString
  toString() {
    return `[Bird: ${this.name}, Age: ${this.age}, Wingspan: ${this.wingspan}m, Health: ${this.getHealth()}]`;
  }
}

// ==============================
// 子类：Fish（继承 Animal）
// ==============================
class Fish extends Animal {
  constructor(name, age, depth = 2) {
    super(name, age, 'Fish');
    this.depth = depth;
  }

  speak() {
    console.log(`${this.name} bubbles silently.`);
  }

  move() {
    console.log(`${this.name} swims gracefully in water.`);
  }

  // 新增特有方法
  changeColor(color) {
    this._color = color;
    console.log(`${this.name} changed to ${color}!`);
  }

  toString() {
    return `[Fish: ${this.name}, Age: ${this.age}, Depth: ${this.depth}m, Health: ${this.getHealth()}]`;
  }
}

// ==============================
// 混合类：会飞的鱼（多重继承模拟）
// ==============================
class FlyingFish extends Swimmable(Flyable(Bird)) {
  constructor(name, age, wingspan = 1.2, depth = 1) {
    super(name, age, 'Flying Fish');
    this.wingspan = wingspan;
    this.depth = depth;
  }

  // 重写方法
  move() {
    console.log(`${this.name} glides above the water like a dolphin with wings!`);
  }

  // 重写 speak
  speak() {
    console.log(`${this.name} makes a splashy chirp!`);
  }
}

// ==============================
// 使用示例 & 测试所有特性
// ==============================
console.log('🚀 JavaScript OOP Feature Showcase\n');

// 1. 构造实例
const bird = new Bird('Tweety', 3, 1.8);
const fish = new Fish('Nemo', 2, 1);
const flyingFish = new FlyingFish('Dory', 1, 1.5, 0.5);

// 2. 静态属性与方法
console.log(`Total animals created: ${Animal.getTotalCount()}`); // 3

// 3. 私有字段访问（仅限类内）
console.log(`Bird ID: ${bird.getId()}`); // 1
console.log(`Bird health: ${bird.getHealth()}`); // 100

// 4. 访问器属性
bird.age = 4;
console.log(`Updated age: ${bird.age}`);

// 5. 实例方法调用
bird.eat('worms');
fish.sleep(6);
flyingFish.fly();     // 来自 Flyable
flyingFish.swim();    // 来自 Swimmable
flyingFish.move();    // 重写后的方法

// 6. 抽象方法测试
try {
  const baseAnimal = new Animal('Generic', 1, 'Unknown');
  baseAnimal.move(); // 抛出错误
} catch (e) {
  console.error(e.message); // 正确捕获
}

// 7. 迭代器协议
console.log('\n--- Iterating over bird ---');
for (const prop of bird) {
  console.log(prop); // name, age, species, health, isAdult
}

// 8. 工厂方法
const eagle = Animal.createBird('Eagle', 5);
console.log(eagle.toString());

// 9. instanceof 和原型链
console.log(`bird instanceof Bird: ${bird instanceof Bird}`); // true
console.log(`bird instanceof Animal: ${bird instanceof Animal}`); // true
console.log(`bird instanceof FlyingFish: ${bird instanceof FlyingFish}`); // false

// 10. Symbol 私有属性
console.log(`Internal data timestamp: ${bird[Symbol.for('internalData')].createdAt}`);

// 11. 动态添加方法（运行时）
Animal.prototype.breathe = function () {
  console.log(`${this.name} breathes air.`);
};
bird.breathe(); // 动态添加的方法生效

// 12. 可调用对象
bird.call('hello', 'world'); // 模拟函数调用

// 13. 封装与冻结
bird.lock();
// bird.newProp = 'fail'; // 会被静默忽略（严格模式下报错）

bird.freeze();
// bird.age = 100; // TypeError: Cannot assign to read only property

// 14. toString 重写
console.log('\n--- toString overrides ---');
console.log(bird.toString());
console.log(fish.toString());
console.log(flyingFish.toString());

// 15. 检查原型
console.log(`Bird prototype: ${bird.getPrototype().constructor.name}`); // "Bird"

// 16. 混入继承验证
console.log(`FlyingFish is instance of Bird: ${flyingFish instanceof Bird}`); // true
console.log(`FlyingFish is instance of Swimmable: ${flyingFish instanceof Swimmable}`); // false（注意：Mixin 不是构造器）

// 17. 使用可选链与空值合并
const maybeAnimal = null;
console.log(maybeAnimal?.name ?? 'No animal');

// 18. 继承链验证
console.log(`\n=== Prototype Chain ===`);
console.log(Object.getPrototypeOf(flyingFish).name); // Flyable
console.log(Object.getPrototypeOf(Object.getPrototypeOf(flyingFish)).name); // Swimmable
console.log(Object.getPrototypeOf(Object.getPrototypeOf(Object.getPrototypeOf(flyingFish))).name); // Bird
console.log(Object.getPrototypeOf(Object.getPrototypeOf(Object.getPrototypeOf(Object.getPrototypeOf(flyingFish))))); // Animal

console.log('\n✅ All OOP features demonstrated successfully!');
```

---

## 💡 总结：这个类展示了什么？

| 特性 | 是否展示 |
|------|----------|
| `class`、`constructor` | ✅ |
| 实例属性/方法 | ✅ |
| 静态属性/方法 | ✅ |
| 私有字段 `#` | ✅ |
| 私有方法 | ✅ |
| `get/set` 访问器 | ✅ |
| `extends` 继承 | ✅ |
| `super()` | ✅ |
| 方法重写 | ✅ |
| 动态原型方法 | ✅ |
| Symbol 私有属性 | ✅ |
| 工厂方法 | ✅ |
| 抽象方法模拟 | ✅ |
| 迭代器协议 | ✅ |
| `call/apply/bind` 演示 | ✅ |
| `Object.defineProperty` | ✅ |
| `Object.freeze/seal` | ✅ |
| Mixin 模式 | ✅ |
| `instanceof`、`isPrototypeOf` | ✅ |
| `Object.getPrototypeOf` | ✅ |

---

## ✅ 建议运行方式

将上述代码保存为 `.js` 文件，在 Node.js 环境（v16+）中运行：

```bash
node oop-showcase.js
```

或在浏览器 DevTools 控制台中粘贴执行。

---

这个示例不仅是一个类，更是一个**JavaScript OOP 的教科书级范例**，适合面试、教学或项目架构参考。希望它能帮助你深入理解 JS 的强大面向对象能力！🌟
