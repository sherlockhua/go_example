// 监听来自popup的消息
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  switch (request.action) {
    case 'saveTabs':
      saveAllTabs(sendResponse);
      return true; // 保持消息通道开放
      
    case 'restoreAllTabs':
      restoreAllTabs(request.index, sendResponse);
      return true;
      
    case 'restoreTab':
      restoreSingleTab(request.groupIndex, request.tabIndex);
      sendResponse({ success: true });
      return true;
  }
});

// 保存所有标签页
function saveAllTabs(sendResponse) {
  // 获取当前窗口的所有标签页
  chrome.tabs.query({ currentWindow: true }, (tabs) => {
    if (tabs.length === 0) {
      sendResponse({ success: false, message: 'No tabs to save' });
      return;
    }
    
    // 提取需要保存的标签页信息
    const tabsToSave = tabs.map(tab => ({
      id: tab.id,
      url: tab.url,
      title: tab.title,
      favIconUrl: tab.favIconUrl
    }));
    
    // 创建标签组
    const tabGroup = {
      timestamp: Date.now(),
      tabs: tabsToSave
    };
    
    // 保存到同步存储
    chrome.storage.sync.get('tabGroups', (result) => {
      const tabGroups = result.tabGroups || [];
      tabGroups.push(tabGroup);
      
      chrome.storage.sync.set({ tabGroups }, () => {
        sendResponse({ success: true });
      });
    });
  });
}

// 恢复所有标签页
function restoreAllTabs(groupIndex, sendResponse) {
  chrome.storage.sync.get('tabGroups', (result) => {
    const tabGroups = result.tabGroups || [];
    
    // 注意：存储的数组是正序，而popup中显示的是倒序
    const actualIndex = tabGroups.length - 1 - groupIndex;
    const group = tabGroups[actualIndex];
    
    if (!group) {
      sendResponse({ success: false, message: 'Group not found' });
      return;
    }
    
    // 恢复每个标签页
    group.tabs.forEach((tab, index) => {
      chrome.tabs.create({ 
        url: tab.url,
        active: index === 0 // 只激活第一个标签页
      });
    });
    
    sendResponse({ success: true });
  });
}

// 恢复单个标签页
function restoreSingleTab(groupIndex, tabIndex) {
  chrome.storage.local.get('tabGroups', (result) => {
    const tabGroups = result.tabGroups || [];
    
    // 注意：存储的数组是正序，而popup中显示的是倒序
    const actualGroupIndex = tabGroups.length - 1 - groupIndex;
    const group = tabGroups[actualGroupIndex];
    
    if (group && group.tabs[tabIndex]) {
      chrome.tabs.create({ 
        url: group.tabs[tabIndex].url,
        active: true
      });
    }
  });
}

// 监听扩展图标点击事件
chrome.action.onClicked.addListener((tab) => {
  chrome.tabs.sendMessage(tab.id, { action: 'toggleSidebar' });
});
