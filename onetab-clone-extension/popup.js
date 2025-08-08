document.addEventListener('DOMContentLoaded', () => {
  const saveTabsBtn = document.getElementById('save-tabs');
  const emptySaveBtn = document.getElementById('empty-save-btn');
  const tabGroupsContainer = document.getElementById('tab-groups-container');
  const emptyState = document.getElementById('empty-state');
  
  // 加载保存的标签组
  loadTabGroups();
  
  // 保存当前所有标签页
  saveTabsBtn.addEventListener('click', saveAllTabs);
  emptySaveBtn.addEventListener('click', saveAllTabs);
  
  // 保存所有标签页的函数
  function saveAllTabs() {
    chrome.runtime.sendMessage({ action: 'saveTabs' }, (response) => {
      if (response.success) {
        loadTabGroups();
      }
    });
  }
  
  // 加载保存的标签组
  function loadTabGroups() {
    chrome.storage.local.get('tabGroups', (result) => {
      const tabGroups = result.tabGroups || [];
      
      // 显示或隐藏空状态
      if (tabGroups.length === 0) {
        emptyState.style.display = 'flex';
        tabGroupsContainer.style.display = 'none';
      } else {
        emptyState.style.display = 'none';
        tabGroupsContainer.style.display = 'block';
        renderTabGroups(tabGroups);
      }
    });
  }
  
  // 渲染标签组
  function renderTabGroups(tabGroups) {
    tabGroupsContainer.innerHTML = '';
    
    // 按时间倒序显示（最新的在前面）
    tabGroups.slice().reverse().forEach((group, groupIndex) => {
      const groupElement = document.createElement('div');
      groupElement.className = 'tab-group';
      
      // 格式化时间
      const date = new Date(group.timestamp);
      const formattedDate = date.toLocaleString();
      
      groupElement.innerHTML = `
        <div class="tab-group-header">
          <div class="tab-group-title">${group.tabs.length} tabs - ${formattedDate}</div>
          <div class="tab-group-actions">
            <button class="tab-group-btn restore-all" data-index="${groupIndex}" title="Restore all tabs">
              <i class="fa fa-window-restore"></i>
            </button>
            <button class="tab-group-btn delete-group" data-index="${groupIndex}" title="Delete group">
              <i class="fa fa-trash"></i>
            </button>
          </div>
        </div>
        <div class="tab-items"></div>
      `;
      
      const tabItemsContainer = groupElement.querySelector('.tab-items');
      
      // 添加标签项
      group.tabs.forEach((tab, tabIndex) => {
        const tabElement = document.createElement('div');
        tabElement.className = 'tab-item';
        
        // 使用favicon或默认图标
        const faviconUrl = tab.favIconUrl || 'https://www.google.com/s2/favicons?domain=unknown';
        
        tabElement.innerHTML = `
          <img src="${faviconUrl}" class="tab-favicon" alt="Favicon">
          <span class="tab-title">${tab.title}</span>
          <button class="tab-group-btn restore-tab" data-group="${groupIndex}" data-tab="${tabIndex}" title="Restore tab">
            <i class="fa fa-external-link"></i>
          </button>
        `;
        
        tabItemsContainer.appendChild(tabElement);
      });
      
      tabGroupsContainer.appendChild(groupElement);
    });
    
    // 添加事件监听器
    addEventListeners();
  }
  
  // 添加事件监听器
  function addEventListeners() {
    // 恢复所有标签页
    document.querySelectorAll('.restore-all').forEach(btn => {
      btn.addEventListener('click', (e) => {
        const groupIndex = parseInt(e.currentTarget.dataset.index);
        chrome.runtime.sendMessage({ 
          action: 'restoreAllTabs', 
          index: groupIndex 
        }, (response) => {
          if (response.success) {
            // 可以选择删除已恢复的组
            // deleteGroup(groupIndex);
          }
        });
      });
    });
    
    // 恢复单个标签页
    document.querySelectorAll('.restore-tab').forEach(btn => {
      btn.addEventListener('click', (e) => {
        const groupIndex = parseInt(e.currentTarget.dataset.group);
        const tabIndex = parseInt(e.currentTarget.dataset.tab);
        
        chrome.runtime.sendMessage({ 
          action: 'restoreTab', 
          groupIndex, 
          tabIndex 
        });
      });
    });
    
    // 删除标签组
    document.querySelectorAll('.delete-group').forEach(btn => {
      btn.addEventListener('click', (e) => {
        const groupIndex = parseInt(e.currentTarget.dataset.index);
        deleteGroup(groupIndex);
      });
    });
  }
  
  // 删除标签组
  function deleteGroup(index) {
    chrome.storage.local.get('tabGroups', (result) => {
      let tabGroups = result.tabGroups || [];
      
      // 从数组中删除指定索引的组（注意：存储的数组是正序，而我们显示的是倒序）
      const actualIndex = tabGroups.length - 1 - index;
      tabGroups.splice(actualIndex, 1);
      
      chrome.storage.local.set({ tabGroups }, () => {
        loadTabGroups();
      });
    });
  }
});
