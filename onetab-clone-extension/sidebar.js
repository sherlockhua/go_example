// 创建侧边栏元素
function createSidebar() {
  // 注入Font Awesome CSS
  const link = document.createElement('link');
  link.rel = 'stylesheet';
  link.href = 'https://cdn.jsdelivr.net/npm/font-awesome@4.7.0/css/font-awesome.min.css';
  document.head.appendChild(link);

  // 创建侧边栏容器
  const sidebar = document.createElement('div');
  sidebar.className = 'sider-ai-sidebar collapsed';
  sidebar.id = 'tabsaver-sidebar';

  // 创建头部
  const header = document.createElement('div');
  header.className = 'sider-ai-header';
  header.innerHTML = `
    <div style="display: flex; align-items: center;">
      <i class="fa fa-th-large" style="margin-right: 8px; color: #2563eb;"></i>
      <span>TabSaver</span>
    </div>
  `;

  // 创建内容区域 - Tab列表
  const content = document.createElement('div');
  content.className = 'sider-ai-content';
  content.id = 'sidebar-content';
  content.innerHTML = `
    <div class="tab-actions">
      <button id="save-all-tabs" class="btn-save-tabs"><i class="fa fa-plus"></i> Save All Tabs</button>
    </div>
    <div class="tab-list-container"></div>
  `;

  // 组合侧边栏
  sidebar.appendChild(header);
  sidebar.appendChild(content);

  // 创建切换按钮
  const toggleBtn = document.createElement('button');
  toggleBtn.className = 'sider-ai-toggle';
  toggleBtn.innerHTML = '<i class="fa fa-chevron-left"></i>';
  toggleBtn.id = 'sidebar-toggle';

  // 添加到页面
  document.body.appendChild(sidebar);
  document.body.appendChild(toggleBtn);

  // 设置切换按钮初始位置
  toggleBtn.style.right = sidebar.classList.contains('collapsed') ? '20px' : '280px';

  // 添加事件监听
  toggleBtn.addEventListener('click', toggleSidebar);
  document.getElementById('save-all-tabs').addEventListener('click', saveAllTabs);

  // 加载保存的标签组
  loadSavedTabs();
}

// 切换侧边栏显示状态
function toggleSidebar() {
  const sidebar = document.getElementById('tabsaver-sidebar');
  const toggleBtn = document.getElementById('sidebar-toggle');
  const isCollapsed = sidebar.classList.toggle('collapsed');
  
  // 切换body的侧边栏展开类
  document.body.classList.toggle('sidebar-expanded', !isCollapsed);
  
  toggleBtn.innerHTML = isCollapsed ? '<i class="fa fa-chevron-right"></i>' : '<i class="fa fa-chevron-left"></i>';
  toggleBtn.style.right = isCollapsed ? '20px' : '280px';
}

// 保存所有打开的标签页
function saveAllTabs() {
  chrome.runtime.sendMessage({ action: 'saveTabs' }, (response) => {
    if (response?.success) loadSavedTabs();
  });
}

// 加载保存的标签组
function loadSavedTabs() {
  console.log('loadSavedTabs: 开始加载标签组数据');
  chrome.storage.sync.get('tabGroups', (data) => {
    console.log('loadSavedTabs: 从storage获取数据', data);
    const container = document.querySelector('.tab-list-container');
    console.log('loadSavedTabs: 找到标签容器', container);

    if (!container) {
      console.error('loadSavedTabs: 未找到.tab-list-container元素');
      return;
    }

    container.innerHTML = '';

    if (!data.tabGroups || data.tabGroups.length === 0) {
      console.log('loadSavedTabs: 没有找到保存的标签组，显示空状态');
      container.innerHTML = `
        <div class="empty-state">
          <i class="fa fa-folder-open-o"></i>
          <p>没有保存的标签组</p>
          <p>点击上方按钮保存当前所有标签页</p>
        </div>
      `;
      return;
    }

    console.log('loadSavedTabs: 找到标签组数据，数量:', data.tabGroups.length);
    data.tabGroups.forEach((group, index) => {
      const groupEl = document.createElement('div');
      groupEl.className = 'tab-group';
      groupEl.innerHTML = `
        <div class="tab-group-header">
          <span class="tab-group-title">${group.title}</span>
          <span class="tab-count">${group.tabs.length}个标签</span>
          <button class="delete-group" data-index="${index}"><i class="fa fa-trash-o"></i></button>
        </div>
        <div class="tab-group-tabs"></div>
      `;

      group.tabs.forEach(tab => {
        const tabEl = document.createElement('div');
        tabEl.className = 'tab-item';
        tabEl.innerHTML = `
          <img class="tab-favicon" src="${tab.favIconUrl || 'icons/icon16.png'}">
          <span class="tab-title">${tab.title}</span>
          <button class="open-tab" data-tab-id="${tab.id}"><i class="fa fa-external-link"></i></button>
        `;
        tabEl.querySelector('.open-tab').addEventListener('click', (e) => {
          e.stopPropagation();
          chrome.tabs.update(parseInt(e.target.dataset.tabId), {active: true});
        });
        groupEl.querySelector('.tab-group-tabs').appendChild(tabEl);
      });

      groupEl.querySelector('.delete-group').addEventListener('click', (e) => {
        e.stopPropagation();
        const tabGroups = data.tabGroups;
        tabGroups.splice(index, 1);
        chrome.storage.sync.set({tabGroups}, loadSavedTabs);
      });

      container.appendChild(groupEl);
    });
  });
}

// 切换AI模型
function changeModel() {
  const modelSelector = document.getElementById('model-selector');
  const selectedModel = modelSelector.value;
  const content = document.getElementById('sidebar-content');

  // 显示模型切换提示
  const modelMessage = document.createElement('div');
  modelMessage.className = 'chat-message ai-message';
  modelMessage.innerHTML = `<div class="message-content">已切换至 <strong>${modelSelector.options[modelSelector.selectedIndex].text}</strong> 模型</div>`;
  content.appendChild(modelMessage);
  content.scrollTop = content.scrollHeight;
}

// 发送消息
function sendMessage() {
  const input = document.querySelector('.chat-input');
  const message = input.value.trim();
  if (!message) return;

  const content = document.getElementById('sidebar-content');
  const modelSelector = document.getElementById('model-selector');
  const selectedModel = modelSelector.options[modelSelector.selectedIndex].text;

  // 添加用户消息
  const userMessage = document.createElement('div');
  userMessage.className = 'chat-message user-message';
  userMessage.innerHTML = `<div class="message-content">${escapeHtml(message)}</div>`;
  content.appendChild(userMessage);

  // 添加加载指示器
  const loadingIndicator = document.createElement('div');
  loadingIndicator.className = 'loading-indicator';
  loadingIndicator.innerHTML = `
    <div class="loading-dot"></div>
    <div class="loading-dot"></div>
    <div class="loading-dot"></div>
  `;
  content.appendChild(loadingIndicator);

  // 清空输入并滚动到底部
  input.value = '';
  content.scrollTop = content.scrollHeight;

  // 模拟AI响应
  simulateAIResponse(message, selectedModel, loadingIndicator);
}

// 模拟AI响应
function simulateAIResponse(message, model, loadingIndicator) {
  const content = document.getElementById('sidebar-content');

  // 根据消息内容生成不同响应
  setTimeout(() => {
    // 移除加载指示器
    content.removeChild(loadingIndicator);

    // 创建AI响应
    const aiMessage = document.createElement('div');
    aiMessage.className = 'chat-message ai-message';

    // 根据不同类型的消息生成不同响应
    if (message.toLowerCase().includes('保存') || message.toLowerCase().includes('tab')) {
      // 保存标签页功能
      saveCurrentTab().then(tab => {
        aiMessage.innerHTML = `
          <div class="message-content">
            <p>已保存标签页:</p>
            <div style="background: white; padding: 10px; border-radius: 6px; margin-top: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.05);">
              <img src="${tab.favIconUrl || 'icons/icon16.png'}" class="tab-favicon">
              <span style="font-weight: 500;">${tab.title}</span>
            </div>
            <p style="margin-top: 8px;">需要我帮您总结此页面内容吗？</p>
          </div>
        `;
        content.appendChild(aiMessage);
        content.scrollTop = content.scrollHeight;
      });
    } else if (message.toLowerCase().includes('总结') || message.toLowerCase().includes('摘要')) {
      // 页面总结功能
      aiMessage.innerHTML = `
        <div class="message-content">
          <p>正在分析当前页面内容...</p>
          <div style="background: #f9fafb; padding: 12px; border-radius: 6px; margin-top: 10px;">
            <h4 style="margin: 0 0 8px 0; font-size: 15px;">页面摘要</h4>
            <p style="margin: 0; font-size: 14px;">根据页面内容，这是一篇关于"${document.title}"的文章。主要讨论了...</p>
            <button style="margin-top: 10px; padding: 6px 12px; background: #f0f9ff; color: #2563eb; border: none; border-radius: 4px; cursor: pointer; font-size: 13px;">
              <i class="fa fa-expand"></i> 查看完整总结
            </button>
          </div>
        </div>
      `;
      content.appendChild(aiMessage);
      content.scrollTop = content.scrollHeight;
    } else if (message.toLowerCase().includes('模型') || message.toLowerCase().includes('切换')) {
      // 模型相关查询
      aiMessage.innerHTML = `
        <div class="message-content">
          <p>当前已选择 <strong>${model}</strong> 模型。您可以通过顶部下拉菜单切换其他AI模型，包括：</p>
          <ul style="margin: 8px 0 0 20px;">
            <li>GPT-4o - 平衡性能与速度的最佳选择</li>
            <li>Claude 3.5 - 长文本处理能力出色</li>
            <li>Gemini 1.5 Pro - 多模态理解能力强</li>
            <li>DeepSeek - 代码理解与生成专家</li>
          </ul>
        </div>
      `;
      content.appendChild(aiMessage);
      content.scrollTop = content.scrollHeight;
    } else {
      // 通用AI响应
      aiMessage.innerHTML = `
        <div class="message-content">
          <p>这是来自 <strong>${model}</strong> 的响应：</p>
          <p style="margin-top: 8px;">${generateGenericResponse(message)}</p>
        </div>
      `;
      content.appendChild(aiMessage);
      content.scrollTop = content.scrollHeight;
    }
  }, 1500 + Math.random() * 1000);
}

// 辅助函数：生成通用响应
function generateGenericResponse(message) {
  const responses = [
    "感谢您的提问！根据您的需求，我建议...",
    "这个问题很有意思。从多个角度分析，我们可以发现...",
    "您提出了一个很好的观点。结合当前上下文，我的理解是...",
    "根据我的分析，最适合您的解决方案是...",
    "让我帮您梳理一下思路：首先...其次...最后..."
  ];
  return responses[Math.floor(Math.random() * responses.length)] + "\n\n这是一个模拟响应，实际应用中会连接真实AI模型API来获取准确回答。"
}

// 辅助函数：HTML转义
function escapeHtml(unsafe) {
  return unsafe
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

// 保存当前标签
function saveCurrentTab() {
  return new Promise((resolve) => {
    chrome.tabs.query({active: true, currentWindow: true}, (tabs) => {
      const currentTab = tabs[0];
      chrome.storage.sync.get('tabGroups', (data) => {
        const tabGroups = data.tabGroups || [];
        const now = new Date();
        const groupTitle = `${now.getFullYear()}-${now.getMonth()+1}-${now.getDate()} ${now.getHours()}:${now.getMinutes()}`;
        tabGroups.unshift({title: groupTitle, tabs: [currentTab]});
        chrome.storage.sync.set({tabGroups}, () => resolve(currentTab));
      });
    });
  });
}

// 直接初始化侧边栏（content_scripts已配置在document_end运行）
createSidebar();