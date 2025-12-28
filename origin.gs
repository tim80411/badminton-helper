/**
 * 目前關於群組的部分是寫死的
 */

const LINE_CHANNEL_ACCESS_TOKEN = '4ULdGSaaLrPmbeoWVdRP8tCgR9ww8ZY1khUvhwi4glx2CnL/O6PuHKUMKhPesxHzop2FG+HAFcIJfnT8+RMQ/pXl9+TY1AUzpX0rmGsYZ8LekQ4sHKDlFeuE83A7KR4hBWJLq/cETqFzBw+PTCuQCgdB04t89/1O/w1cDnyilFU=';
const TEST_LINE_USER_ID = 'Ca759a35126b4618fa3e30b0562ad43e7';  // 測試
const LINE_USER_ID = 'Cb5ebf211e42d7565efea8a6a0a281ff0';  // 個人或群組的 ID，用於發送消息

function doPost(e) {
  const eventData = JSON.parse(e.postData.contents)
  console.log('Received event data: ' + JSON.stringify(eventData));
  // postToSheet(JSON.stringify(eventData));

  if (eventData.events && eventData.events.length > 0) {
    eventData.events.forEach(function (event) {
      if (event.type === 'join') {
        if (event.source.type === 'group') {
          const groupId = event.source.groupId;
          getAndSaveGroupInfo(groupId);
          sendWelcomeMessageReply(event.replyToken);
        }
      }
      else if (event.type === 'join' && event.source.type === 'room') {
        const roomId = event.source.roomId;
        saveRoomInfo(roomId);
        sendWelcomeMessageReply(event.replyToken);
      }
      else if (event.type === 'postback') {
        handleUserPostback(event);
      }
    });
  }

  // 回應 LINE 平台，確認我們已收到事件
  return true
}


/**
 * 處理用戶發送的消息（修改版 - 傳遞真實 timestamp）
 * @param {Object} event - LINE 消息事件物件
 */
function handleUserPostback(event) {
  // 只處理文字消息
  const data = event.postback.data || ''
  const dataObj = Object.fromEntries(data.split(',').map(item => item.split('=')));
  const action = dataObj.action || 'register'; // 新增 action 參數
  const activityId = dataObj.activityId || `N-${Utilities.getUuid()}`;
  const messageText = dataObj.message || '';
  const registrationId = dataObj.registrationId || ''; // 新增 registrationId 參數

  const replyToken = event.replyToken;
  // 關鍵修改：取得 LINE webhook 的真實 timestamp
  const eventTimestamp = new Date(event.timestamp);

  if (event.source.type === 'group') {
    const groupId = event.source.groupId;
    const userId = event.source.userId;

    logMessage(groupId, messageText, event.source.userId);

    if (action === 'cancel') {
      // 處理取消報名
      handleCancellation(groupId, registrationId, userId, replyToken);
    } else if (action === 'weekly_list') {
      if (!LINE_ADMINS.includes(userId)) {
        console.log('非admin無法觸發週名單')
        return;
      }
      handleWeeklyListRequest(event.source, replyToken, activityId);
    } else if (messageText.includes('臨打') || messageText.includes('季繳')) {
      // 修改：傳遞真實的 timestamp
      handleRegistration(groupId, activityId, messageText, event.source.userId, replyToken, eventTimestamp);
    }
  }
  // 處理其他類型的消息來源
  else if (event.source.type === 'user') {
  }
  else if (event.source.type === 'room') {
  }
}

/**
 * 處理報名指令，例如 "臨打 +1" 或 "季繳 +1"（修改版 - 使用真實 timestamp）
 * @param {string} sourceId - 群組ID或聊天室ID
 * @param {string} activityId - 活動ID
 * @param {string} message - 用戶發送的消息
 * @param {string} userId - 用戶ID
 * @param {string} replyToken - 回覆令牌
 * @param {Date} eventTimestamp - LINE webhook 的真實 timestamp
 */
function handleRegistration(sourceId, activityId, message, userId, replyToken, eventTimestamp) {
  getGroupMemberProfile(sourceId, userId)
    .then(function (profile) {
      // 修改：傳遞真實的 timestamp
      recordRegistrationWithReply(sourceId, activityId, message, profile, replyToken, eventTimestamp);
    })
    .catch(function (error) {
      console.log('Error getting user profile: ' + error);
    });
}


/**
 * 獲取並保存群組信息
 * @param {string} groupId - 群組ID
 */
function getAndSaveGroupInfo(groupId) {
  // 使用 LINE API 獲取群組摘要信息
  console.log('groupId', groupId)
  const url = 'https://api.line.me/v2/bot/group/' + groupId + '/summary';

  // 設定 HTTP 請求頭
  const headers = {
    'Authorization': 'Bearer ' + LINE_CHANNEL_ACCESS_TOKEN
  };

  // 設定 HTTP 請求選項
  const options = {
    'method': 'get',
    'headers': headers,
    'muteHttpExceptions': true  // 使錯誤不中斷腳本運行
  };

  try {
    // 發送 HTTP 請求
    const response = UrlFetchApp.fetch(url, options);
    const responseCode = response.getResponseCode();

    if (responseCode === 200) {
      // 成功獲取群組信息
      const groupInfo = JSON.parse(response.getContentText());

      // 將群組信息保存到 Sheet
      saveGroupInfoToSheet(groupInfo);

      console.log('Successfully retrieved group info: ' + JSON.stringify(groupInfo));
      return groupInfo;
    } else {
      // 記錄錯誤響應
      console.log('Failed to get group info. Response code: ' + responseCode);
      console.log('Response content: ' + response.getContentText());
      return null;
    }
  } catch (error) {
    // 記錄錯誤
    console.log('Error getting group info: ' + error);
    return null;
  }
}

/**
 * 處理本週名單請求
 * @param {Object} source - 消息來源物件
 * @param {string} replyToken - 回覆令牌
 * @param {string} activityId - 活動ID
 */
function handleWeeklyListRequest(source, replyToken, activityId) {
  try {
    const sourceId = source.groupId || source.roomId || source.userId;
    const weeklyRegistrations = getWeeklyRegistrations(sourceId, activityId);

    // 創建名單顯示的 Flex Message
    const weeklyListMessage = createWeeklyListMessage(weeklyRegistrations, activityId);

    replyMessage(replyToken, weeklyListMessage);
  } catch (error) {
    console.log('Error handling weekly list request: ' + error);
    const errorMessage = {
      type: "text",
      text: "取得本週名單時發生錯誤，請稍後再試。"
    };
    replyMessage(replyToken, errorMessage);
  }
}

/**
 * 獲取本週的報名資料
 * @param {string} sourceId - 群組/聊天室/用戶ID
 * @param {string} activityId - 活動ID
 * @returns {Object} 包含季繳和臨打報名資料的物件
 */
function getWeeklyRegistrations(sourceId, activityId) {
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  const sheet = ss.getSheetByName('Registrations');

  if (!sheet) {
    return { seasonTicket: [], casualPlay: [] };
  }

  const data = sheet.getDataRange().getValues();
  const headers = data[0];

  // 找到各欄位的索引
  const timestampIndex = headers.indexOf('Timestamp');
  const groupIdIndex = headers.indexOf('Group ID');
  const activityIdIndex = headers.indexOf('ActivityId');
  const userIdIndex = headers.indexOf('User ID');
  const userNameIndex = headers.indexOf('User Name');
  const registrationTypeIndex = headers.indexOf('Registration Type');
  const courtAssignmentIndex = 7; // H欄，索引7
  const cancelledIndex = headers.indexOf('isCancelled');

  // 獲取本週的開始和結束時間（週一到週日）
  const { weekStart, weekEnd } = getCurrentWeekRange();

  const seasonTicket = [];
  const casualPlay = [];

  // 遍歷所有報名記錄
  for (let i = 1; i < data.length; i++) {
    const row = data[i];
    const timestamp = new Date(row[timestampIndex]);
    const registrationType = row[registrationTypeIndex];
    const isCancelled = row[cancelledIndex] === true || row[cancelledIndex] === 'TRUE';
    const currentSourceId = row[groupIdIndex];
    const currentActivityId = row[activityIdIndex];

    // 過濾條件：
    // 1. 時間在本週範圍內
    // 2. 來源ID匹配
    // 3. 活動ID匹配
    // 4. 未取消報名
    if (timestamp >= weekStart &&
      timestamp <= weekEnd &&
      currentSourceId === sourceId &&
      currentActivityId === activityId &&
      !isCancelled) {

      const registrationData = {
        userId: row[userIdIndex],
        userName: row[userNameIndex],
        timestamp: timestamp,
        court: row[courtAssignmentIndex] || '', // H欄的場地資訊
        registrationType: registrationType
      };

      if (registrationType === '季繳') {
        seasonTicket.push(registrationData);
      } else if (registrationType === '臨打') {
        casualPlay.push(registrationData);
      }
    }
  }

  // 按報名時間排序（先報名的在前面）
  seasonTicket.sort((a, b) => a.timestamp - b.timestamp);
  casualPlay.sort((a, b) => a.timestamp - b.timestamp);

  return { seasonTicket, casualPlay };
}

/**
 * 獲取當前週的時間範圍（週一到週日）
 * @returns {Object} 包含 weekStart 和 weekEnd 的物件
 */
function getCurrentWeekRange() {
  const now = new Date();
  const currentDay = now.getDay(); // 0=週日, 1=週一, ..., 6=週六

  // 計算本週一的日期
  // 如果今天是週一(1)，則本週一就是今天，daysToMonday = 0
  // 如果今天是週二(2)，則本週一是昨天，daysToMonday = -1
  // 如果今天是週日(0)，則本週一是6天前，daysToMonday = -6
  let daysToMonday;
  if (currentDay === 0) {
    daysToMonday = -6; // 週日的話，本週一是6天前
  } else {
    daysToMonday = 1 - currentDay; // 其他情況：1-當前日期數字
  }

  // 設定本週一（週的開始）
  const weekStart = new Date(now);
  weekStart.setDate(now.getDate() + daysToMonday);
  weekStart.setHours(0, 0, 0, 0);

  // 設定本週日（週的結束）
  const weekEnd = new Date(weekStart);
  weekEnd.setDate(weekStart.getDate() + 6);
  weekEnd.setHours(23, 59, 59, 999);

  return { weekStart, weekEnd };
}

/**
 * 創建本週名單的 Flex Message
 * @param {Object} registrations - 包含季繳和臨打報名資料的物件
 * @returns {Object} Flex Message 物件
 */
function createWeeklyListMessage(registrations) {
  const { seasonTicket, casualPlay } = registrations;
  const isSaturday = new Date().getDay() === 6; // 判斷是否為週六

  // 建立名單內容
  const contents = [];

  // 添加標題
  contents.push({
    "type": "text",
    "text": "本週報名名單 📋",
    "weight": "bold",
    "size": "xl",
    "margin": "md"
  });

  // 季繳名單
  if (seasonTicket.length > 0) {
    contents.push({
      "type": "text",
      "text": "季繳：",
      "weight": "bold",
      "size": "md",
      "color": "#333333",
      "margin": "lg"
    });

    seasonTicket.forEach((reg, index) => {
      let displayText = `${index + 1}. ${reg.userName}`;

      // 週六才顯示場地資訊
      if (isSaturday && reg.court) {
        displayText += `：${reg.court}`;
      }

      contents.push({
        "type": "text",
        "text": displayText,
        "size": "sm",
        "color": "#666666",
        "margin": "xs"
      });
    });
  }

  // 臨打名單
  if (casualPlay.length > 0) {
    const startIndex = seasonTicket.length + 1;

    contents.push({
      "type": "text",
      "text": "臨打：",
      "weight": "bold",
      "size": "md",
      "color": "#333333",
      "margin": "lg"
    });

    casualPlay.forEach((reg, index) => {
      let displayText = `${startIndex + index}. ${reg.userName}`;

      // 週六才顯示場地資訊
      if (isSaturday && reg.court) {
        displayText += `：${reg.court}`;
      }

      contents.push({
        "type": "text",
        "text": displayText,
        "size": "sm",
        "color": "#666666",
        "margin": "xs"
      });
    });
  }

  // 如果沒有任何報名
  if (seasonTicket.length === 0 && casualPlay.length === 0) {
    contents.push({
      "type": "text",
      "text": "本週還沒有人報名唷！快來報名吧 🏸",
      "size": "md",
      "color": "#999999",
      "margin": "lg",
      "align": "center"
    });
  }

  // 添加提示文字
  if (!isSaturday) {
    contents.push({
      "type": "text",
      "text": "💡 場地分配將在週六公布",
      "size": "xs",
      "color": "#999999",
      "margin": "lg",
      "align": "center"
    });
  }

  return {
    "type": "flex",
    "altText": "本週羽球報名名單",
    "contents": {
      "type": "bubble",
      "body": {
        "type": "box",
        "layout": "vertical",
        "contents": contents,
        "spacing": "sm"
      },
      "footer": {
        "type": "box",
        "layout": "vertical",
        "spacing": "sm",
        "contents": [
          {
            "type": "text",
            "text": "想報名的話請點選上面的按鈕唷！",
            "size": "xs",
            "color": "#aaaaaa",
            "align": "center"
          }
        ]
      }
    }
  };
}

/**
 * 處理取消報名的邏輯
 * @param {string} groupId - 群組ID
 * @param {string} registrationId - 報名記錄ID
 * @param {string} userId - 用戶ID
 * @param {string} replyToken - 回覆令牌
 */
function handleCancellation(groupId, registrationId, userId, replyToken) {
  try {
    // 驗證用戶是否有權限取消此報名
    const validationResult = validateCancellationPermission(registrationId, userId);

    if (!validationResult.isValid) {
      if (validationResult.needUserName) {
        getGroupMemberProfile(groupId, userId)
          .then(function (profile) {
            const errorMessage = {
              type: "text",
              text: `${profile.displayName} 壞壞，不要取消別人的預約👀`
            };
            replyMessage(replyToken, errorMessage);
          })
          .catch(function (error) {
            console.log('Error getting user profile: ' + error);
            const errorMessage = {
              type: "text",
              text: "你壞壞，不要取消別人的預約👀"
            };
            replyMessage(replyToken, errorMessage);
          });
      } else {
        // 其他錯誤直接回傳
        const errorMessage = {
          type: "text",
          text: validationResult.errorMessage
        };
        replyMessage(replyToken, errorMessage);
      }
      return;
    }

    // 執行取消操作
    const cancellationResult = cancelRegistration(registrationId);

    if (cancellationResult.success) {
      const successMessage = {
        type: "text",
        text: `✅ ${cancellationResult.userName} 的 ${cancellationResult.registrationType} 報名已成功取消！`
      };
      replyMessage(replyToken, successMessage);
    } else {
      const errorMessage = {
        type: "text",
        text: "取消報名時發生錯誤，請稍後再試。"
      };
      replyMessage(replyToken, errorMessage);
    }
  } catch (error) {
    console.log('Error handling cancellation: ' + error);
    const errorMessage = {
      type: "text",
      text: "系統錯誤，無法處理取消請求。"
    };
    replyMessage(replyToken, errorMessage);
  }
}

/**
 * 執行取消報名操作
 * @param {string} registrationId - 報名記錄ID
 * @returns {Object} 操作結果
 */
function cancelRegistration(registrationId) {
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  const sheet = ss.getSheetByName('Registrations');
  const data = sheet.getDataRange().getValues();
  const headers = data[0];

  let cancelledColumnIndex = headers.indexOf('isCancelled');
  if (cancelledColumnIndex === -1) {
    cancelledColumnIndex = headers.length;
  }

  // 尋找要取消的記錄
  for (let i = 1; i < data.length; i++) {
    const row = data[i];
    const rowRegistrationId = generateRegistrationId(row[0], row[3]);

    if (rowRegistrationId === registrationId) {
      // 標記為已取消
      sheet.getRange(i + 1, cancelledColumnIndex + 1).setValue(true);

      return {
        success: true,
        userName: row[4], // User Name 欄位
        registrationType: row[5] // Registration Type 欄位
      };
    }
  }

  return {
    success: false
  };
}

/**
 * 基本驗證用戶是否有權限取消指定的報名
 * @param {string} registrationId - 報名記錄ID
 * @param {string} userId - 用戶ID
 * @returns {Object} 驗證結果
 */
function validateCancellationPermission(registrationId, userId) {
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  const sheet = ss.getSheetByName('Registrations');

  if (!sheet) {
    return {
      isValid: false,
      needUserName: false,
      errorMessage: "找不到報名記錄。"
    };
  }

  const data = sheet.getDataRange().getValues();
  const headers = data[0];

  // 確保 isCancelled 欄位存在
  let cancelledColumnIndex = headers.indexOf('isCancelled');
  if (cancelledColumnIndex === -1) {
    // 如果不存在，新增這個欄位
    sheet.getRange(1, headers.length + 1).setValue('isCancelled');
    cancelledColumnIndex = headers.length;
  }

  // 尋找對應的報名記錄
  for (let i = 1; i < data.length; i++) {
    const row = data[i];
    const rowRegistrationId = generateRegistrationId(row[0], row[3]); // timestamp + userId

    if (rowRegistrationId === registrationId) {
      // 檢查是否為同一用戶
      if (row[3] !== userId) { // User ID 欄位
        return {
          isValid: false,
          needUserName: true, // 標記需要取得用戶名稱
          errorMessage: "" // 錯誤訊息將在取得用戶名稱後產生
        };
      }

      // 檢查是否已經取消過
      if (row[cancelledColumnIndex] === true || row[cancelledColumnIndex] === 'TRUE') {
        return {
          isValid: false,
          needUserName: false,
          errorMessage: "這個報名已經被取消過囉，請大人放過我的伺服器"
        };
      }

      return {
        isValid: true,
        needUserName: false,
        rowIndex: i + 1 // +1 因為 Google Sheets 從 1 開始計數
      };
    }
  }

  return {
    isValid: false,
    needUserName: false,
    errorMessage: "找不到對應的報名記錄。是不是管理員在壞壞！"
  };
}


/**
 * 生成報名記錄的唯一ID
 * @param {Date|string} timestamp - 時間戳
 * @param {string} userId - 用戶ID
 * @returns {string} 唯一ID
 */
function generateRegistrationId(timestamp, userId) {
  // 將時間戳轉換為字串並結合用戶ID
  const timeStr = new Date(timestamp).getTime().toString();
  return `${timeStr}-${userId}`;
}



/**
 * 將群組信息保存到 Sheet
 * @param {Object} groupInfo - 群組信息物件
 */
function saveGroupInfoToSheet(groupInfo) {
  // 獲取活動的 Spreadsheet
  const ss = SpreadsheetApp.getActiveSpreadsheet();

  // 尋找或創建 "Groups" 表格
  let sheet = ss.getSheetByName('Groups');
  if (!sheet) {
    sheet = ss.insertSheet('Groups');
    // 設置表頭
    sheet.appendRow(['Group ID', 'Group Name', 'Picture URL', 'Join Date', 'Member Count']);
  }

  // 檢查此群組 ID 是否已存在
  const data = sheet.getDataRange().getValues();
  let groupExists = false;
  let rowIndex = -1;

  for (let i = 1; i < data.length; i++) {  // 從第 2 行開始（跳過表頭）
    if (data[i][0] === groupInfo.groupId) {
      groupExists = true;
      rowIndex = i + 1;  // +1 因為 getValues() 返回的 0 索引對應於 Sheet 的第 1 行
      break;
    }
  }

  // 當前日期時間
  const currentDate = new Date();

  if (groupExists) {
    // 更新現有群組信息
    sheet.getRange(rowIndex, 2).setValue(groupInfo.groupName);
    sheet.getRange(rowIndex, 3).setValue(groupInfo.pictureUrl || 'No picture');
    sheet.getRange(rowIndex, 5).setValue(groupInfo.memberCount);
  } else {
    // 添加新群組
    sheet.appendRow([
      groupInfo.groupId,
      groupInfo.groupName,
      groupInfo.pictureUrl || 'No picture',
      currentDate,
      groupInfo.memberCount
    ]);
  }
}

/**
 * 保存聊天室信息（因為 LINE API 不提供獲取聊天室名稱的接口，所以只保存 ID）
 * @param {string} roomId - 聊天室 ID
 */
function saveRoomInfo(roomId) {
  // 獲取活動的 Spreadsheet
  const ss = SpreadsheetApp.getActiveSpreadsheet();

  // 尋找或創建 "Rooms" 表格
  let sheet = ss.getSheetByName('Rooms');
  if (!sheet) {
    sheet = ss.insertSheet('Rooms');
    // 設置表頭
    sheet.appendRow(['Room ID', 'Join Date']);
  }

  // 檢查此聊天室 ID 是否已存在
  const data = sheet.getDataRange().getValues();
  let roomExists = false;

  for (let i = 1; i < data.length; i++) {  // 從第 2 行開始（跳過表頭）
    if (data[i][0] === roomId) {
      roomExists = true;
      break;
    }
  }

  // 如果聊天室不存在，添加新記錄
  if (!roomExists) {
    // 當前日期時間
    const currentDate = new Date();
    sheet.appendRow([roomId, currentDate]);
  }
}

/**
 * 使用 reply 方法發送歡迎消息
 * @param {string} replyToken - 回覆令牌
 */
function sendWelcomeMessageReply(replyToken) {
  const message = {
    type: "text",
    text: "感謝將我加入群組！我是一個羽球活動報名機器人，可以協助管理報名和發送定時訊息。點選活動中的季繳/臨打可以進行報名"
  };

  replyMessage(replyToken, message);
}

/**
 * 獲取用戶個人資料
 * @param {string} userId - 用戶 ID
 * @returns {Promise} - 包含用戶個人資料的 Promise
 */
function getUserProfile(userId) {
  return new Promise(function (resolve, reject) {
    const url = 'https://api.line.me/v2/bot/profile/' + userId;

    // 設定 HTTP 請求頭
    const headers = {
      'Authorization': 'Bearer ' + LINE_CHANNEL_ACCESS_TOKEN
    };

    // 設定 HTTP 請求選項
    const options = {
      'method': 'get',
      'headers': headers,
      'muteHttpExceptions': true
    };

    try {
      // 發送 HTTP 請求
      const response = UrlFetchApp.fetch(url, options);
      const responseCode = response.getResponseCode();

      if (responseCode === 200) {
        // 成功獲取用戶個人資料
        const profile = JSON.parse(response.getContentText());
        resolve(profile);
      } else {
        // 記錄錯誤響應
        console.log('Failed to get user profile. Response code: ' + responseCode);
        reject(new Error('Failed to get user profile. Response code: ' + responseCode));
      }
    } catch (error) {
      // 記錄錯誤
      console.log('Error getting user profile: ' + error);
      reject(error);
    }
  });
}

/**
 * 獲取 LINE 群組中特定成員的個人資料
 * @param {string} groupId - 群組 ID
 * @param {string} userId - 用戶 ID
 * @returns {Promise} - 包含用戶個人資料的 Promise
 */
function getGroupMemberProfile(groupId, userId) {
  return new Promise((resolve, reject) => {
    const url = `https://api.line.me/v2/bot/group/${groupId}/member/${userId}`;

    // 設定 HTTP 請求頭
    const headers = {
      'Authorization': `Bearer ${LINE_CHANNEL_ACCESS_TOKEN}`
    };

    // 設定 HTTP 請求選項
    const options = {
      'method': 'get',
      'headers': headers,
      'muteHttpExceptions': true
    };

    try {
      // 發送 HTTP 請求
      const response = UrlFetchApp.fetch(url, options);
      const responseCode = response.getResponseCode();

      if (responseCode === 200) {
        // 成功獲取用戶個人資料
        const profile = JSON.parse(response.getContentText());
        resolve(profile);
      } else {
        // 記錄錯誤響應
        console.log(`Failed to get group member profile. Response code: ${responseCode}`);
        reject(new Error(`Failed to get group member profile. Response code: ${responseCode}`));
      }
    } catch (error) {
      // 記錄錯誤
      console.log(`Error getting group member profile: ${error}`);
      reject(error);
    }
  });
}

/**
 * 使用 reply 記錄報名信息到表格並回覆用戶（修改版 - 使用真實 timestamp）
 * @param {string} groupId - 群組 ID
 * @param {string} activityId - 活動ID
 * @param {string} message - 報名消息
 * @param {Object} profile - 用戶個人資料
 * @param {string} replyToken - 回覆令牌
 * @param {Date} eventTimestamp - LINE webhook 的真實 timestamp
 */
function recordRegistrationWithReply(groupId, activityId, message, profile, replyToken, eventTimestamp) {
  // 獲取活動的 Spreadsheet
  const ss = SpreadsheetApp.getActiveSpreadsheet();

  // 尋找或創建 "Registrations" 表格
  let sheet = ss.getSheetByName('Registrations');
  if (!sheet) {
    sheet = ss.insertSheet('Registrations');
    // 設置表頭
    sheet.appendRow(['Timestamp', 'Group ID', 'ActivityId', 'User ID', 'User Name', 'Registration Type', 'Message']);
  }

  // 關鍵修改：使用傳入的 eventTimestamp 而不是 new Date()
  // 這確保了報名順序是按照用戶實際點擊的時間，而不是伺服器處理的時間
  const registrationTimestamp = eventTimestamp;

  // 確定報名類型
  let registrationType = '';
  if (message.includes('臨打')) {
    registrationType = '臨打';
  } else if (message.includes('季繳')) {
    registrationType = '季繳';
  } else {
    registrationType = '其他';
  }

  // 添加新報名記錄 - 使用真實的 timestamp
  sheet.appendRow([
    registrationTimestamp,  // 使用 LINE 的真實時間
    groupId,
    activityId,
    profile.userId,
    profile.displayName,
    registrationType,
    message,
    null,
    false
  ]);

  // 生成 registrationId 時也使用相同的 timestamp，確保一致性
  const registrationId = generateRegistrationId(registrationTimestamp, profile.userId);

  // 創建包含取消按鈕的回覆訊息
  const confirmMessage = createRegistrationConfirmMessage(
    profile.displayName,
    registrationType,
    registrationId
  );

  replyMessage(replyToken, confirmMessage);
}

/**
 * 創建包含取消按鈕的報名確認訊息
 * 使用你提供的新樣式
 * @param {string} userName - 用戶名稱
 * @param {string} registrationType - 報名類型
 * @param {string} registrationId - 報名記錄ID
 * @returns {Object} LINE 訊息物件
 */
function createRegistrationConfirmMessage(userName, registrationType, registrationId) {
  return {
    "type": "flex",
    "altText": `${userName} 的 ${registrationType} 報名確認`,
    "contents": {
      "type": "bubble",
      "body": {
        "type": "box",
        "layout": "vertical",
        "contents": [
          {
            "type": "box",
            "layout": "vertical",
            "margin": "lg",
            "spacing": "sm",
            "contents": [
              {
                "type": "text",
                "text": `已記錄 ${userName} 的 ${registrationType} 報名！想變卦請點按鈕退場 👇`,
                "color": "#272727",
                "size": "md",
                "wrap": true,
                "margin": "xs"
              }
            ]
          }
        ]
      },
      "footer": {
        "type": "box",
        "layout": "vertical",
        "spacing": "sm",
        "contents": [
          {
            "type": "button",
            "style": "primary",
            "height": "sm",
            "action": {
              "type": "postback",
              "label": "取消報名",
              "data": `action=cancel,registrationId=${registrationId},userName=${userName}`,
              "displayText": `${userName} 申請取消報名！`
            },
            "color": "#CE0000"
          },
          {
            "type": "box",
            "layout": "vertical",
            "contents": [],
            "margin": "sm"
          }
        ],
        "flex": 0
      }
    }
  };
}

/**
 * 保留原有 recordRegistration 功能，以防需要
 * @param {string} groupId - 群組 ID
 * @param {string} message - 報名消息
 * @param {Object} profile - 用戶個人資料
 */
function recordRegistration(groupId, message, profile) {
  // 獲取活動的 Spreadsheet
  const ss = SpreadsheetApp.getActiveSpreadsheet();

  // 尋找或創建 "Registrations" 表格
  let sheet = ss.getSheetByName('Registrations');
  if (!sheet) {
    sheet = ss.insertSheet('Registrations');
    // 設置表頭
    sheet.appendRow(['Timestamp', 'Group ID', 'User ID', 'User Name', 'Registration Type', 'Message']);
  }

  // 當前日期時間
  const currentDate = new Date();

  // 確定報名類型
  let registrationType = '';
  if (message.includes('臨打')) {
    registrationType = '臨打';
  } else if (message.includes('季繳')) {
    registrationType = '季繳';
  } else {
    registrationType = '其他';
  }

  // 添加新報名記錄
  sheet.appendRow([
    currentDate,
    groupId,
    profile.userId,
    profile.displayName,
    registrationType,
    message
  ]);

  // 可選：發送確認消息回群組
  const confirmMessage = {
    type: "text",
    text: `已記錄 ${profile.displayName} 的 ${registrationType} 報名。謝謝！`
  };

  sendMessage(groupId, confirmMessage);
}

/**
 * 記錄聊天消息（可選修改 - 也可以使用真實 timestamp）
 * @param {string} sourceId - 消息來源 ID（群組或聊天室）
 * @param {string} message - 消息內容
 * @param {string} userId - 用戶 ID
 * @param {Date} eventTimestamp - 可選：LINE webhook 的真實 timestamp
 */
function logMessage(sourceId, message, userId, eventTimestamp = null) {
  // 獲取活動的 Spreadsheet
  const ss = SpreadsheetApp.getActiveSpreadsheet();

  // 尋找或創建 "Messages" 表格
  let sheet = ss.getSheetByName('Messages');
  if (!sheet) {
    sheet = ss.insertSheet('Messages');
    // 設置表頭
    sheet.appendRow(['Timestamp', 'Source ID', 'User ID', 'Message']);
  }

  // 使用傳入的 timestamp 或當前時間
  const messageTimestamp = eventTimestamp || new Date();

  // 添加新消息記錄
  sheet.appendRow([
    messageTimestamp,
    sourceId,
    userId,
    message
  ]);
}

/**
 * 使用 reply 方法發送消息
 * @param {string} replyToken - 回覆令牌
 * @param {Object} message - 消息物件
 */
function replyMessage(replyToken, message) {
  const url = 'https://api.line.me/v2/bot/message/reply';

  // 設定 HTTP 請求頭
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer ' + LINE_CHANNEL_ACCESS_TOKEN
  };

  // 構建要發送的數據
  const postData = {
    'replyToken': replyToken,
    'messages': [message],
    'notificationDisabled': true
  };

  // 設定 HTTP 請求選項
  const options = {
    'method': 'post',
    'headers': headers,
    'payload': JSON.stringify(postData),
    'muteHttpExceptions': true
  };

  try {
    // 發送 HTTP 請求
    const response = UrlFetchApp.fetch(url, options);
    const responseCode = response.getResponseCode();

    if (responseCode === 200) {
      // 消息發送成功
      console.log("Reply message sent successfully.");
      return true;
    } else {
      // 記錄錯誤響應
      console.log('Failed to send reply message. Response code: ' + responseCode);
      console.log('Response content: ' + response.getContentText());
      return false;
    }
  } catch (error) {
    // 記錄錯誤
    console.log("Error sending reply message: " + error);
    return false;
  }
}

/**
 * 發送消息到指定目標 (push 方法，保留原有功能)
 * @param {string} targetId - 目標 ID（用戶、群組或聊天室）
 * @param {Object} message - 消息物件
 */
function sendMessage(targetId, message) {
  const url = 'https://api.line.me/v2/bot/message/push';

  // 設定 HTTP 請求頭
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer ' + LINE_CHANNEL_ACCESS_TOKEN
  };

  // 構建要發送的數據
  const postData = {
    'to': targetId,
    'messages': [message]
  };

  // 設定 HTTP 請求選項
  const options = {
    'method': 'post',
    'headers': headers,
    'payload': JSON.stringify(postData),
    'muteHttpExceptions': true
  };

  try {
    // 發送 HTTP 請求
    const response = UrlFetchApp.fetch(url, options);
    const responseCode = response.getResponseCode();

    if (responseCode === 200) {
      // 消息發送成功
      console.log("Message sent successfully.");
      return true;
    } else {
      // 記錄錯誤響應
      console.log('Failed to send message. Response code: ' + responseCode);
      console.log('Response content: ' + response.getContentText());
      return false;
    }
  } catch (error) {
    // 記錄錯誤
    console.log("Error sending message: " + error);
    return false;
  }
}

/**
 * 建立一個可以由觸發器調用的函數，用於定時發送消息
 * 這個函數將使用 Flex Message 結構
 */
function sendScheduledMessage() {
  // 創建 Flex Message 結構
  const flexMessage = createFlexMessage();

  // 從 Sheet 獲取所有需要發送消息的群組
  // const groups = getAllGroups();

  // 發送 Flex Message 到每個群組
  sendFlexMessageToTarget(LINE_USER_ID, flexMessage);
}

/**
 * 獲取所有儲存的群組 ID
 * @returns {Array} - 群組 ID 數組
 */
function getAllGroups() {
  // 獲取活動的 Spreadsheet
  const ss = SpreadsheetApp.getActiveSpreadsheet();

  // 尋找 "Groups" 表格
  const sheet = ss.getSheetByName('Groups');
  if (!sheet) {
    return [];
  }

  // 獲取所有數據
  const data = sheet.getDataRange().getValues();

  // 跳過表頭，獲取所有群組 ID
  const groupIds = [];
  for (let i = 1; i < data.length; i++) {
    if (data[i][0]) {  // 確保 ID 存在
      groupIds.push(data[i][0]);
    }
  }

  return groupIds;
}

/**
 * 發送 Flex Message 到指定目標
 * @param {string} targetId - 目標 ID（用戶、群組或聊天室）
 * @param {Object} flexContent - Flex Message 內容
 */
function sendFlexMessageToTarget(targetId, flexContent) {
  const url = 'https://api.line.me/v2/bot/message/push';

  // 設定 HTTP 請求頭
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer ' + LINE_CHANNEL_ACCESS_TOKEN
  };

  // 構建完整的 Flex Message 格式
  const message = {
    'type': 'flex',
    'altText': '羽球活動報名',
    'contents': flexContent
  };

  // 構建要發送的數據
  const postData = {
    'to': targetId,
    'messages': [message]
  };

  // 設定 HTTP 請求選項
  const options = {
    'method': 'post',
    'headers': headers,
    'payload': JSON.stringify(postData),
    'muteHttpExceptions': true
  };

  try {
    // 發送 HTTP 請求
    const response = UrlFetchApp.fetch(url, options);
    const responseCode = response.getResponseCode();

    if (responseCode === 200) {
      // 消息發送成功
      console.log(`Flex message sent successfully to ${targetId}.`);
      return true;
    } else {
      // 記錄錯誤響應
      console.log(`Failed to send flex message to ${targetId}. Response code: ${responseCode}`);
      console.log('Response content: ' + response.getContentText());
      return false;
    }
  } catch (error) {
    // 記錄錯誤
    console.log(`Error sending flex message to ${targetId}: ${error}`);
    return false;
  }
}

function createPauseByRaceFlexMsg() {
  return {
    "type": "bubble",
    "body": {
      "type": "box",
      "layout": "vertical",
      "contents": [
        {
          "type": "text",
          "text": "本週羽球報名暫停通知",
          "weight": "bold",
          "size": "xl"
        },
        {
          "type": "box",
          "layout": "vertical",
          "margin": "lg",
          "spacing": "sm",
          "contents": [
            {
              "type": "text",
              "text": "本週羽球報名因舉行比賽暫停一次"
            }
          ]
        }
      ]
    }
  }
}

/**
 * 創建 Flex Message 結構
 * 使用您提供的模板
 */
function createFlexMessage() {
  const uuid = Utilities.getUuid()
  const today = new Date();
  const todayStr = Utilities.formatDate(today, "Asia/Taipei", "yyyyMMdd");

  // 保留原有的 Flex Message 結構，在 footer 部分加入新按鈕
  const flexContent = {
    "type": "bubble",
    "hero": {
      "type": "image",
      "size": "full",
      "aspectRatio": "20:13",
      "aspectMode": "cover",
      "action": {
        "type": "uri",
        "uri": "https://line.me/"
      },
      "url": "https://i.imgur.com/sUrFITq.png"
    },
    "body": {
      "type": "box",
      "layout": "vertical",
      "contents": [
        {
          "type": "text",
          "text": "phah kiû lah 🏸",
          "weight": "bold",
          "size": "xl"
        },
        {
          "type": "box",
          "layout": "vertical",
          "margin": "lg",
          "spacing": "sm",
          "contents": [
            {
              "type": "box",
              "layout": "baseline",
              "spacing": "sm",
              "contents": [
                {
                  "type": "text",
                  "text": "Place",
                  "color": "#aaaaaa",
                  "size": "sm",
                  "flex": 1
                },
                {
                  "type": "text",
                  "text": `${SETTING.place}`,
                  "wrap": true,
                  "color": "#666666",
                  "size": "sm",
                  "flex": 5
                }
              ]
            },
            {
              "type": "box",
              "layout": "baseline",
              "spacing": "sm",
              "contents": [
                {
                  "type": "text",
                  "text": "Time",
                  "color": "#aaaaaa",
                  "size": "sm",
                  "flex": 1
                },
                {
                  "type": "text",
                  "text": `${SETTING.time}`,
                  "wrap": true,
                  "color": "#666666",
                  "size": "sm",
                  "flex": 5
                }
              ]
            },
            {
              "type": "box",
              "layout": "baseline",
              "contents": [
                {
                  "type": "text",
                  "text": "Price",
                  "color": "#aaaaaa",
                  "size": "sm",
                  "flex": 1
                },
                {
                  "type": "text",
                  "text": `${SETTING.price}`,
                  "wrap": true,
                  "color": "#666666",
                  "size": "sm",
                  "flex": 5
                }
              ],
              "spacing": "sm"
            },
            {
              "type": "box",
              "layout": "baseline",
              "contents": [
                {
                  "type": "text",
                  "text": "Quota",
                  "color": "#aaaaaa",
                  "size": "sm",
                  "flex": 1
                },
                {
                  "type": "text",
                  "text": `${SETTING.quota}`,
                  "wrap": true,
                  "color": "#666666",
                  "size": "sm",
                  "flex": 5
                }
              ],
              "spacing": "sm"
            },
            {
              "type": "text",
              "text": "＊想報名幾位就按幾次按鈕，例如：報名臨打 2 位，就按兩次「臨打」",
              "color": "#666666",
              "size": "sm",
              "wrap": true,
              "margin": "md"
            },
            {
              "type": "text",
              "text": "⚠️ 臨時有事需取消請提前告知",
              "color": "#EA0000",
              "size": "sm"
            }
          ]
        }
      ]
    },
    "footer": {
      "type": "box",
      "layout": "vertical",
      "spacing": "sm",
      "contents": [
        {
          "type": "button",
          "style": "primary",
          "height": "sm",
          "action": {
            "type": "postback",
            "label": "半年繳 / 季繳",
            "data": `activityId=${todayStr}-${uuid},message=季繳+1`,
            "displayText": "季繳 +1"
          }
        },
        {
          "type": "button",
          "style": "secondary",
          "height": "sm",
          "action": {
            "type": "postback",
            "label": "臨打",
            "data": `activityId=${todayStr}-${uuid},message=臨打+1`,
            "displayText": "臨打 +1"
          }
        },
        {
          "type": "button",
          "style": "link",
          "height": "sm",
          "action": {
            "type": "postback",
            "label": "本週名單",
            "data": `action=weekly_list,activityId=${todayStr}-${uuid}`,
            "displayText": "查看本週名單"
          },
          "color": "#17c950"
        },
        {
          "type": "box",
          "layout": "vertical",
          "contents": [],
          "margin": "sm"
        }
      ],
      "flex": 0
    }
  };

  return flexContent;
}


/**
 * 設置定時觸發
 * 這個函數用於設置定時任務，可以在編輯器中手動運行一次來啟用定時器
 */
function setUpTrigger() {
  // 刪除所有現有的觸發器，以避免重複
  const triggers = ScriptApp.getProjectTriggers();
  for (let i = 0; i < triggers.length; i++) {
    ScriptApp.deleteTrigger(triggers[i]);
  }

  // 創建一個新的時間驅動的觸發器，例如每週五早上 9 點運行（為週六的活動做準備）
  ScriptApp.newTrigger('sendScheduledMessage')
    .timeBased()
    .onWeekDay(ScriptApp.WeekDay.TUESDAY)
    .atHour(13)
    .create();

  console.log("Trigger has been set up successfully!");
}

/**
 * 輔助函數：為了在後台表單 H 欄輸入場地資訊提供便利
 * 這個函數可以幫助管理員快速查看需要分配場地的報名
 */
function getRegistrationsForCourtAssignment() {
  const registrations = getWeeklyRegistrations(LINE_USER_ID); // 或其他群組ID
  const { seasonTicket, casualPlay } = registrations;

  console.log('=== 本週報名名單（供場地分配參考）===');
  console.log('季繳報名：');
  seasonTicket.forEach((reg, index) => {
    console.log(`${index + 1}. ${reg.userName} (${reg.userId}) - 報名時間: ${reg.timestamp}`);
  });

  console.log('\n臨打報名：');
  casualPlay.forEach((reg, index) => {
    console.log(`${seasonTicket.length + index + 1}. ${reg.userName} (${reg.userId}) - 報名時間: ${reg.timestamp}`);
  });
}

/**
 * 測試本週名單功能
 * @param {string} activityId - 活動ID（可選，用於測試特定活動）
 */
function testWeeklyList(activityId) {
  // 如果沒有提供 activityId，使用當天的格式生成一個測試用的
  if (!activityId) {
    const today = new Date();
    const todayStr = Utilities.formatDate(today, "Asia/Taipei", "yyyyMMdd");
    const uuid = Utilities.getUuid();
    activityId = `${todayStr}-${uuid}`;
    activityId = '20250805-63d32192-5775-45bb-a96e-1c67ed35f1d5'
  }

  const registrations = getWeeklyRegistrations(LINE_USER_ID, activityId);
  const message = createWeeklyListMessage(registrations, activityId);

  console.log('本週名單測試結果：');
  console.log(`測試活動ID: ${activityId}`);
  console.log(JSON.stringify(message, null, 2));

  // 實際發送測試訊息
  sendFlexMessageToTarget(TEST_LINE_USER_ID, message.contents);
}

/**
 * 測試函數
 * 可以手動運行此函數來測試消息發送是否正常工作
 */
function testSendMessage() {
  // 創建 Flex Message
  const flexMessage = createFlexMessage();

  // 只發送到指定的測試 ID（可以是群組 ID 或個人 ID）
  sendFlexMessageToTarget(TEST_LINE_USER_ID, flexMessage);
}

/**
 * 測試取消功能的函數
 */
function testCancelFunction() {
  // 這個函數可以用來測試取消功能是否正常運作
  const testRegistrationId = generateRegistrationId(new Date(), 'test-user-id');
  console.log('Generated test registration ID:', testRegistrationId);

  // 可以在這裡加入更多測試邏輯
}

/**
 * 將數據保存到 Sheet
 * @param {string} data - 要保存的數據
 */
function postToSheet(data) {
  var ss = SpreadsheetApp.getActiveSpreadsheet();
  var sheet = ss.getSheetByName('EventLog');

  // 如果 sheet 不存在，創建一個
  if (!sheet) {
    sheet = ss.insertSheet('EventLog');
    sheet.appendRow(['Timestamp', 'Event Data']);
  }

  // 添加新行，包含時間戳和事件數據
  sheet.appendRow([new Date(), data]);
}