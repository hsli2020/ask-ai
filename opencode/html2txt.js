function htmlToText(html) {
    const temp = document.createElement('div');
    temp.innerHTML = html;

    let result = '';
    let listDepth = 0;
    let listType = [];

    function traverse(node) {
        if (node.nodeType === Node.TEXT_NODE) {
            result += node.textContent;
            return;
        }

        if (node.nodeType !== Node.ELEMENT_NODE) {
            return;
        }

        const tagName = node.tagName.toLowerCase();

        // 处理列表
        if (tagName === 'ul' || tagName === 'ol') {
            listDepth++;
            listType.push(tagName);
            for (const child of node.childNodes) {
                traverse(child);
            }
            listDepth--;
            listType.pop();
            if (!result.endsWith('\n')) result += '\n';
            return;
        }

        // 处理列表项
        if (tagName === 'li') {
            const bullet = listType[listType.length - 1] === 'ol' ? '1. ' : '• ';
            result += '\n' + '  '.repeat(listDepth - 1) + bullet;
            for (const child of node.childNodes) {
                traverse(child);
            }
            return;
        }

        // 处理表格
        if (tagName === 'table') {
            result += '\n';
            for (const child of node.childNodes) {
                traverse(child);
            }
            result += '\n';
            return;
        }

        if (tagName === 'tr') {
            let cells = [];
            for (const child of node.childNodes) {
                if (child.tagName && (child.tagName.toLowerCase() === 'td' || child.tagName.toLowerCase() === 'th')) {
                    let cellText = '';
                    const originalResult = result;
                    result = '';
                    traverse(child);
                    cellText = result;
                    result = originalResult;
                    cells.push(cellText.trim());
                }
            }
            result += '| ' + cells.join(' | ') + ' |\n';
            return;
        }

        // 块级元素
        const blockElements = ['p', 'div', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'blockquote', 'pre', 'code'];
        const isBlock = blockElements.includes(tagName);

        if (isBlock && result && !result.endsWith('\n')) {
            result += '\n';
        }

        // 标题格式
        if (tagName.startsWith('h')) {
            const level = parseInt(tagName[1]);
            result += '#'.repeat(level) + ' ';
        }

        // 引用
        if (tagName === 'blockquote') {
            result += '> ';
        }

        // 代码块
        if (tagName === 'pre') {
            result += '```\n';
        }

        // 处理子节点
        for (const child of node.childNodes) {
            traverse(child);
        }

        // 块级元素后
        if (tagName === 'pre') {
            result += '\n```\n';
        }

        if (isBlock && !result.endsWith('\n') && tagName !== 'pre') {
            result += '\n';
        }

        if (tagName === 'br') {
            result += '\n';
        }
    }

    traverse(temp);

    // 清理多余空白
    result = result
        .replace(/[ \t]+/g, ' ')
        .replace(/\n[ \t]+/g, '\n')
        .replace(/[ \t]+\n/g, '\n')
        .replace(/\n{3,}/g, '\n\n');

    return result.trim();
}
