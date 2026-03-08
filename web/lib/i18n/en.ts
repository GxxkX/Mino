import type { Dictionary } from './zh-CN';

const en: Dictionary = {
  // Common
  loading: 'Loading...',
  cancel: 'Cancel',
  confirm: 'Confirm',
  delete: 'Delete',
  save: 'Save',
  close: 'Close',
  add: 'Add',
  search: 'Search...',
  noResults: 'No results found',
  untitled: 'Untitled',

  // Nav / Sidebar
  nav: {
    schedule: 'Schedule',
    memories: 'Memories',
    tasks: 'Tasks',
    audio: 'Audio',
    chat: 'Chat',
    extensions: 'Extensions',
    settings: 'Settings',
    navigation: 'Navigation',
  },

  // Header
  header: {
    search: 'Search...',
    recording: 'Recording',
    record: 'Record',
    conversations: 'Conversations',
    memoriesLabel: 'Memories',
    back: 'Back',
  },

  // Auth / Login
  auth: {
    welcome: 'Welcome back',
    subtitle: 'Sign in to your personal AI assistant',
    username: 'Username',
    usernamePlaceholder: 'Enter username',
    password: 'Password',
    passwordPlaceholder: 'Enter password',
    loginButton: 'Sign in',
    loggingIn: 'Signing in...',
    loginError: 'Invalid username or password',
  },

  // Dashboard
  dashboard: {
    title: 'Daily Review',
    description: 'Your personal data overview',
    weekConversations: 'This Week',
    memoryPoints: 'Memories',
    pendingTasks: 'Pending',
    completed: 'Completed',
    recentMemories: 'Recent Memories',
    recentConversations: 'Recent Conversations',
    pendingTasksSection: 'Pending Tasks',
    viewAll: 'View all',
  },

  // Memories
  memories: {
    title: 'Memories',
    countDescription: ((count: number) => `${count} memories`) as Dictionary['memories']['countDescription'],
    searchPlaceholder: 'Search memories...',
    empty: 'No memories yet',
    all: 'All',
    insight: 'Insight',
    fact: 'Fact',
    preference: 'Preference',
    event: 'Event',
  },

  // Tasks
  tasks: {
    title: 'Tasks',
    description: ((pending: number, inProgress: number, completed: number) =>
      `${pending} pending / ${inProgress} in progress / ${completed} completed`) as Dictionary['tasks']['description'],
    total: 'Total',
    pending: 'Pending',
    inProgress: 'In Progress',
    completed: 'Completed',
    cancelled: 'Cancelled',
    all: 'All',
    empty: 'No tasks yet',
    newTask: 'New Task',
    priorityLow: 'Low',
    priorityMedium: 'Medium',
    priorityHigh: 'High',
    dueDate: ((date: string) => `Due ${date}`) as Dictionary['tasks']['dueDate'],
  },

  // Create Task Dialog
  createTask: {
    title: 'New Task',
    titleLabel: 'Title',
    titlePlaceholder: 'Enter task title...',
    titleRequired: 'Please enter a task title',
    descLabel: 'Description',
    descOptional: '(optional)',
    descPlaceholder: 'Add more details...',
    priority: 'Priority',
    dueDateLabel: 'Due Date',
    dueDateOptional: '(optional)',
    creating: 'Creating...',
    create: 'Create Task',
    createError: 'Failed to create, please try again',
  },

  // Audio / Recordings
  audio: {
    title: 'Recordings',
    description: ((count: number, minutes: number) => `${count} recordings / ${minutes} min`) as Dictionary['audio']['description'],
    totalRecordings: 'Total',
    minutes: 'min',
    processed: 'Processed',
    searchPlaceholder: 'Search recordings...',
    empty: 'No recordings yet',
    deleteTitle: 'Delete Recording',
    deleteConfirm: ((title: string) => `Are you sure you want to delete "${title}"? This cannot be undone.`) as Dictionary['audio']['deleteConfirm'],
    deleting: 'Deleting...',
    deleteAudioFile: 'Also delete audio file',
  },

  // Chat
  chat: {
    title: 'Chat',
    newChat: 'New Chat',
    deleteSession: 'Delete session',
    emptySessions: 'No sessions yet',
    welcomeTitle: 'How can I help you?',
    welcomeDesc: 'I can help you review past conversations, answer questions, and organize your thoughts.',
    startChat: 'Start a new chat',
    sendPlaceholder: 'Type a message...',
    startMessage: 'Send a message to start',
    thinking: 'Mino is thinking...',
    errorReply: 'Sorry, an error occurred. Please try again later.',
    sources: 'Sources',
  },

  // Extensions
  extensions: {
    title: 'Extensions',
    description: 'Manage and configure your extensions',
    empty: 'No extensions yet. Add one below.',
    enabled: 'Enabled',
    installed: 'Installed',
    addExtension: 'Add Extension',
    namePlaceholder: 'Extension name',
    descPlaceholder: 'Description (optional)',
    adding: 'Adding...',
    confirmAdd: 'Confirm',
    configure: 'Configure',
  },

  // Settings
  settings: {
    title: 'Settings',
    general: 'General',
    language: 'Language',
    cloudSync: 'Cloud Sync',
    cloudSettings: 'Cloud Settings',
    cloudEnabled: 'Enabled',
    cloudDisabled: 'Disabled',
    config: 'Configuration',
    llmProvider: 'LLM Provider',
    security: 'Security',
    changePassword: 'Change Password',
    signOut: 'Sign Out',
    user: 'User',
    recording: 'Recording',
    recordingGain: 'Recording Gain',
    gainStandard: 'Standard',
    gainMute: 'Mute',
    gainWarning: 'Gain above 100% may cause audio clipping',
  },

  // Security Settings
  securitySettings: {
    title: 'Account Security',
    changeUsernameTitle: 'Change Username',
    changeUsernameDesc: 'Update your login username',
    newUsername: 'New Username',
    newUsernamePlaceholder: 'At least 2 characters',
    usernameTooShort: 'Username must be at least 2 characters',
    usernamePassword: 'Current Password',
    usernamePasswordPlaceholder: 'Enter password to confirm',
    usernameSuccess: 'Username changed, signing out...',
    usernameError: 'Failed to change username',
    submitUsername: 'Change Username',
    changingUsername: 'Changing...',
    changePasswordTitle: 'Change Password',
    changePasswordDesc: 'Change your password regularly for better security',
    currentPassword: 'Current Password',
    currentPasswordPlaceholder: 'Enter current password',
    newPassword: 'New Password',
    newPasswordPlaceholder: 'At least 6 characters',
    confirmPassword: 'Confirm Password',
    confirmPasswordPlaceholder: 'Re-enter new password',
    passwordTooShort: 'Password must be at least 6 characters',
    passwordMismatch: 'Passwords do not match',
    success: 'Password changed, signing out...',
    error: 'Failed to change password',
    changing: 'Changing...',
    submit: 'Change Password',
  },

  // Cloud Settings
  cloudSettings: {
    title: 'Cloud Settings',
    syncTitle: 'Cloud Sync',
    syncDesc: 'Automatically sync data to the cloud',
    storageConfig: 'Object Storage (MinIO)',
    testConnection: 'Test Connection',
    saveConfig: 'Save Config',
    saving: 'Saving...',
    saved: 'Configuration saved',
    saveError: 'Failed to save, please try again',
    // MinIO
    minioEndpoint: 'Endpoint',
    minioEndpointPlaceholder: 'localhost:9000',
    minioAccessKey: 'Access Key',
    minioSecretKey: 'Secret Key',
    minioSecure: 'HTTPS',
    minioRegion: 'Region',
    minioRegionPlaceholder: 'us-east-1',
    minioPublicUrl: 'Public URL',
    minioPublicUrlPlaceholder: 'https://cdn.example.com',
    // PostgreSQL
    postgresql: 'PostgreSQL',
    dbHost: 'Host',
    dbPort: 'Port',
    dbName: 'Database',
    dbUser: 'Username',
    dbPassword: 'Password',
    dbSslMode: 'SSL Mode',
    // Redis
    redis: 'Redis',
    redisHost: 'Host',
    redisPort: 'Port',
    redisPassword: 'Password',
    redisDb: 'DB Number',
    // Milvus
    milvus: 'Milvus Vector DB',
    milvusHost: 'Host',
    milvusPort: 'Port',
    milvusUser: 'Username',
    milvusPassword: 'Password',
    milvusDbName: 'Database',
    // Typesense
    typesense: 'Typesense Search',
    typesenseHost: 'Host',
    typesensePort: 'Port',
    typesenseApiKey: 'API Key',
  },

  // Recording Banner
  recordingBanner: {
    paused: 'Paused',
    recording: 'Recording',
    resume: 'Resume recording',
    pause: 'Pause recording',
    stop: 'Stop recording',
    waiting: 'Waiting for voice input...',
  },

  // Confirm Dialog
  confirmDialog: {
    confirm: 'Confirm',
    cancel: 'Cancel',
    close: 'Close',
  },

  // LLM Settings
  llmSettings: {
    title: 'LLM Configuration',
    provider: 'Provider',
    providerDesc: 'Select your LLM service provider',
    apiKey: 'API Key',
    apiKeyPlaceholder: 'Enter API Key',
    baseUrl: 'Base URL',
    baseUrlPlaceholder: 'Custom API endpoint (optional)',
    model: 'Model',
    modelPlaceholder: 'Enter model name',
    embeddingModel: 'Embedding Model',
    embeddingModelPlaceholder: 'Enter embedding model name',
    save: 'Save Config',
    saving: 'Saving...',
    saved: 'Configuration saved',
    saveError: 'Failed to save, please try again',
    testConnection: 'Test Connection',
    testing: 'Testing...',
    testSuccess: 'Connection successful',
    testError: 'Connection failed',
    openai: 'OpenAI',
    anthropic: 'Anthropic',
    ollama: 'Ollama',
    openaiDesc: 'GPT-4o, GPT-4, GPT-3.5, etc.',
    anthropicDesc: 'Claude 4, Claude 3.5, etc.',
    ollamaDesc: 'Locally deployed open-source models',
    currentConfig: 'Current Config',
    modelConfig: 'Model Config',
  },

  // MCP Settings
  mcpSettings: {
    title: 'MCP Settings',
    protocol: 'MCP Protocol',
    protocolDesc: 'Enable Model Context Protocol',
    servers: 'MCP Servers',
    add: 'Add',
    unnamed: 'Unnamed',
    name: 'Name',
    namePlaceholder: 'Server name',
    type: 'Type',
    typePlaceholder: 'filesystem, http...',
    endpoint: 'Endpoint',
    endpointPlaceholder: 'http://localhost:8080',
    delete: 'Delete',
    saveConfig: 'Save Config',
  },

  // Audio Visualizer
  audioVisualizer: {
    active: 'Recording - audio frequency visualization',
    paused: 'Recording paused',
    inactive: 'Recording not started',
  },
};

export default en;
