const zhCN = {
  // Common
  loading: '加载中...',
  cancel: '取消',
  confirm: '确认',
  delete: '删除',
  save: '保存',
  close: '关闭',
  add: '添加',
  search: '搜索...',
  noResults: '未找到相关结果',
  untitled: '无标题',

  // Nav / Sidebar
  nav: {
    schedule: '日程',
    memories: '记忆',
    tasks: '任务',
    audio: '录音',
    chat: '对话',
    extensions: '扩展',
    settings: '设置',
    navigation: '导航',
  },

  // Header
  header: {
    search: '搜索...',
    recording: '录音中',
    record: '录音',
    conversations: '录音 / 对话',
    memoriesLabel: '记忆',
    back: '返回',
  },

  // Auth / Login
  auth: {
    welcome: '欢迎回来',
    subtitle: '登录你的个人 AI 助手',
    username: '账号',
    usernamePlaceholder: '输入账号',
    password: '密码',
    passwordPlaceholder: '输入密码',
    loginButton: '登录',
    loggingIn: '登录中...',
    loginError: '账号或密码错误',
  },

  // Dashboard
  dashboard: {
    title: '日程回顾',
    description: '你的个人数据概览',
    weekConversations: '本周对话',
    memoryPoints: '记忆点',
    pendingTasks: '待办任务',
    completed: '已完成',
    recentMemories: '最近记忆',
    recentConversations: '最近对话',
    pendingTasksSection: '待办任务',
    viewAll: '查看全部',
  },

  // Memories
  memories: {
    title: '记忆',
    countDescription: (count: number) => `共 ${count} 条记忆`,
    searchPlaceholder: '搜索记忆...',
    empty: '暂无记忆',
    all: '全部',
    insight: '见解',
    fact: '事实',
    preference: '偏好',
    event: '事件',
  },

  // Tasks
  tasks: {
    title: '任务',
    description: (pending: number, inProgress: number, completed: number) =>
      `${pending} 待处理 / ${inProgress} 进行中 / ${completed} 已完成`,
    total: '总计',
    pending: '待处理',
    inProgress: '进行中',
    completed: '已完成',
    cancelled: '已取消',
    all: '全部',
    empty: '暂无任务',
    newTask: '新建任务',
    priorityLow: '低',
    priorityMedium: '中',
    priorityHigh: '高',
    dueDate: (date: string) => `截止 ${date}`,
  },

  // Create Task Dialog
  createTask: {
    title: '新建任务',
    titleLabel: '标题',
    titlePlaceholder: '输入任务标题...',
    titleRequired: '请输入任务标题',
    descLabel: '描述',
    descOptional: '（可选）',
    descPlaceholder: '添加更多细节...',
    priority: '优先级',
    dueDateLabel: '截止日期',
    dueDateOptional: '（可选）',
    creating: '创建中...',
    create: '创建任务',
    createError: '创建失败，请重试',
  },

  // Audio / Recordings
  audio: {
    title: '录音',
    description: (count: number, minutes: number) => `${count} 条录音 / ${minutes} 分钟`,
    totalRecordings: '总录音',
    minutes: '分钟',
    processed: '已处理',
    searchPlaceholder: '搜索录音...',
    empty: '暂无录音',
    deleteTitle: '删除录音',
    deleteConfirm: (title: string) => `确定要删除「${title}」吗？此操作无法撤销。`,
    deleting: '删除中...',
    deleteAudioFile: '同时删除录音音频文件',
  },

  // Chat
  chat: {
    title: '对话',
    newChat: '新对话',
    deleteSession: '删除会话',
    emptySessions: '暂无会话',
    welcomeTitle: '有什么可以帮你的？',
    welcomeDesc: '我可以帮你回顾历史对话、回答问题、整理思路。',
    startChat: '开始新对话',
    sendPlaceholder: '输入消息...',
    startMessage: '发送消息开始对话',
    thinking: 'Mino 正在思考...',
    errorReply: '抱歉，发生了错误，请稍后重试。',
    sources: '引用来源',
  },

  // Extensions
  extensions: {
    title: '扩展',
    description: '管理和配置你的扩展程序',
    empty: '暂无扩展，点击下方添加',
    enabled: '已启用',
    installed: '已安装',
    addExtension: '添加扩展',
    namePlaceholder: '扩展名称',
    descPlaceholder: '描述（可选）',
    adding: '添加中...',
    confirmAdd: '确认添加',
    configure: '配置',
  },

  // Settings
  settings: {
    title: '设置',
    general: '通用',
    language: '语言',
    cloudSync: '云同步',
    cloudSettings: '云端设置',
    cloudEnabled: '已启用',
    cloudDisabled: '已禁用',
    config: '配置',
    llmProvider: 'LLM 提供商',
    security: '安全',
    changePassword: '修改密码',
    signOut: '退出登录',
    user: '用户',
    recording: '录音',
    recordingGain: '录音增益',
    gainStandard: '标准',
    gainMute: '静音',
    gainWarning: '增益大于 100% 时录音可能产生削波失真',
  },

  // Security Settings
  securitySettings: {
    title: '账户安全',
    changePasswordTitle: '修改密码',
    changePasswordDesc: '定期更换密码以保障账户安全',
    currentPassword: '当前密码',
    currentPasswordPlaceholder: '输入当前密码',
    newPassword: '新密码',
    newPasswordPlaceholder: '至少 6 个字符',
    confirmPassword: '确认新密码',
    confirmPasswordPlaceholder: '再次输入新密码',
    passwordTooShort: '密码长度至少 6 个字符',
    passwordMismatch: '两次输入的密码不一致',
    success: '密码修改成功',
    error: '密码修改失败',
    changing: '修改中...',
    submit: '修改密码',
  },

  // Cloud Settings
  cloudSettings: {
    title: '云端设置',
    syncTitle: '云同步',
    syncDesc: '自动同步数据到云端',
    storageConfig: '对象存储 (MinIO)',
    testConnection: '测试连接',
    saveConfig: '保存配置',
    saving: '保存中...',
    saved: '配置已保存',
    saveError: '保存失败，请重试',
    // MinIO
    minioEndpoint: '端点',
    minioEndpointPlaceholder: 'localhost:9000',
    minioAccessKey: 'Access Key',
    minioSecretKey: 'Secret Key',
    minioSecure: 'HTTPS',
    minioRegion: '区域',
    minioRegionPlaceholder: 'us-east-1',
    minioPublicUrl: '公网 URL',
    minioPublicUrlPlaceholder: 'https://cdn.example.com',
    // PostgreSQL
    postgresql: 'PostgreSQL',
    dbHost: '主机',
    dbPort: '端口',
    dbName: '数据库名',
    dbUser: '用户名',
    dbPassword: '密码',
    dbSslMode: 'SSL 模式',
    // Redis
    redis: 'Redis',
    redisHost: '主机',
    redisPort: '端口',
    redisPassword: '密码',
    redisDb: '数据库编号',
    // Milvus
    milvus: 'Milvus 向量数据库',
    milvusHost: '主机',
    milvusPort: '端口',
    milvusUser: '用户名',
    milvusPassword: '密码',
    milvusDbName: '数据库名',
    // Typesense
    typesense: 'Typesense 搜索',
    typesenseHost: '主机',
    typesensePort: '端口',
    typesenseApiKey: 'API Key',
  },

  // Recording Banner
  recordingBanner: {
    paused: '已暂停',
    recording: '正在录音',
    resume: '继续录音',
    pause: '暂停录音',
    stop: '停止录音',
    waiting: '等待语音输入...',
  },

  // Confirm Dialog
  confirmDialog: {
    confirm: '确认',
    cancel: '取消',
    close: '关闭',
  },

  // LLM Settings
  llmSettings: {
    title: 'LLM 配置',
    provider: '提供商',
    providerDesc: '选择 LLM 服务提供商',
    apiKey: 'API Key',
    apiKeyPlaceholder: '输入 API Key',
    baseUrl: 'Base URL',
    baseUrlPlaceholder: '自定义 API 端点（可选）',
    model: '模型',
    modelPlaceholder: '输入模型名称',
    embeddingModel: 'Embedding 模型',
    embeddingModelPlaceholder: '输入 Embedding 模型名称',
    save: '保存配置',
    saving: '保存中...',
    saved: '配置已保存',
    saveError: '保存失败，请重试',
    testConnection: '测试连接',
    testing: '测试中...',
    testSuccess: '连接成功',
    testError: '连接失败',
    openai: 'OpenAI',
    zhipu: '智谱 AI',
    ollama: 'Ollama',
    openaiDesc: 'GPT-4o, GPT-4, GPT-3.5 等',
    zhipuDesc: 'GLM-4, GLM-3 等',
    ollamaDesc: '本地部署的开源模型',
    currentConfig: '当前配置',
    modelConfig: '模型配置',
  },

  // MCP Settings
  mcpSettings: {
    title: 'MCP 设置',
    protocol: 'MCP 协议',
    protocolDesc: '启用 Model Context Protocol',
    servers: 'MCP 服务器',
    add: '添加',
    unnamed: '未命名',
    name: '名称',
    namePlaceholder: '服务器名称',
    type: '类型',
    typePlaceholder: 'filesystem, http...',
    endpoint: '端点',
    endpointPlaceholder: 'http://localhost:8080',
    delete: '删除',
    saveConfig: '保存配置',
  },

  // Audio Visualizer
  audioVisualizer: {
    active: '正在录音 - 音频频率可视化',
    paused: '录音已暂停',
    inactive: '录音未开始',
  },
};

export default zhCN;

// Use widened string types so other locales can provide different values
export type Dictionary = {
  [K in keyof typeof zhCN]: typeof zhCN[K] extends string
    ? string
    : typeof zhCN[K] extends (...args: infer A) => string
      ? (...args: A) => string
      : {
          [P in keyof typeof zhCN[K]]: typeof zhCN[K][P] extends string
            ? string
            : typeof zhCN[K][P] extends (...args: infer A2) => string
              ? (...args: A2) => string
              : typeof zhCN[K][P];
        };
};
