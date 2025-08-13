# 查看网页的源代码时，发现很多元素都有以data-开头的属性，但是标准中又查不到这些属性的含义，这些属性到底是什么？

`data-` 开头的属性是 **HTML5 自定义数据属性（Custom Data Attributes）**，它们是 HTML5 标准的一部分，专门用于在 HTML 元素上存储自定义数据。

## 基本概念

### 1. 语法规范
- 必须以 `data-` 开头
- 只能包含小写字母、数字、连字符（-）和点（.）
- 不能包含大写字母或特殊字符

```html
<!-- 正确的用法 -->
<div data-user-id="123" data-user-name="张三" data-role="admin"></div>
<div data-product-price="29.99" data-product-category="electronics"></div>

<!-- 错误的用法 -->
<div data-UserId="123">        <!-- 包含大写字母 -->
<div data_user_id="123">       <!-- 使用下划线而不是连字符 -->
```

### 2. 在 JavaScript 中访问

#### 使用 dataset API（推荐）
```javascript
const element = document.querySelector('[data-user-id]');
console.log(element.dataset.userId);        // "123"
console.log(element.dataset.userName);      // "张三"
console.log(element.dataset.role);          // "admin"

// 设置数据
element.dataset.status = "active";
```

**注意**：`dataset` 会自动将连字符转换为驼峰命名：
- `data-user-id` → `dataset.userId`
- `data-user-name` → `dataset.userName`

#### 使用 getAttribute/setAttribute
```javascript
const element = document.querySelector('[data-user-id]');
console.log(element.getAttribute('data-user-id')); // "123"
element.setAttribute('data-user-status', 'active');
```

### 3. 在 CSS 中使用

可以通过属性选择器来样式化元素：

```css
/* 选择具有特定 data 属性的元素 */
[data-role="admin"] {
    background-color: #ffeb3b;
}

[data-product-category="electronics"] {
    border: 2px solid blue;
}

/* 使用属性值作为内容 */
[data-tooltip]:hover::after {
    content: attr(data-tooltip);
    /* 其他样式 */
}
```

## 主要用途

### 1. 存储页面相关数据
```html
<!-- 电商网站商品卡片 -->
<div class="product-card" 
     data-product-id="5678" 
     data-price="199.99" 
     data-in-stock="true"
     data-category="books">
    <h3>JavaScript权威指南</h3>
    <p>¥199.99</p>
</div>
```

### 2. 前端框架/库的集成
```html
<!-- Vue.js -->
<div v-for="item in items" :data-item-id="item.id">
    {{ item.name }}
</div>

<!-- React (通过 dangerouslySetInnerHTML 或其他方式) -->
<div data-react-component="UserProfile" data-user-id="123"></div>
```

### 3. 行为配置
```html
<!-- 模态框配置 -->
<button data-modal-target="user-modal" data-modal-animation="fade">
    打开用户信息
</button>

<div id="user-modal" data-modal-backdrop="true" data-modal-keyboard="true">
    <!-- 模态框内容 -->
</div>
```

### 4. A/B 测试和分析
```html
<button data-experiment="button-color-test" data-variant="blue">
    立即购买
</button>
```

## 优势

1. **语义化**：让 HTML 更具语义性
2. **无侵入性**：不会影响页面渲染
3. **标准化**：是 HTML5 官方标准
4. **兼容性好**：现代浏览器都支持
5. **易于维护**：数据与 DOM 元素紧密关联

## 注意事项

1. **不要存储敏感数据**：所有数据对用户可见
2. **避免存储大量数据**：会影响页面性能
3. **命名要清晰**：使用有意义的属性名
4. **考虑可访问性**：屏幕阅读器通常忽略这些属性

## 浏览器支持

- Chrome 8+
- Firefox 6+
- Safari 6+
- Edge 12+
- IE 11+（部分支持）

这些 `data-` 属性是现代 Web 开发中非常实用的工具，让你能够在不污染全局命名空间的情况下，在 HTML 元素上存储和传递自定义数据。
