const messages = {
    'zh-CN': {
        toolCount: '{count} 个工具', refresh: '刷新', settings: '设置', addTool: '+ 添加工具',
        empty: '还没有管理的工具', addFirst: '添加第一个工具', operationFailed: '操作失败',
        operationSuccess: '{id} {operation} {version} 成功', confirmUninstall: '确认卸载 {id}？',
        confirmRemove: '从清单移除 {id}（不影响已安装的二进制）？', noVersions: '{id}：没有可用版本',
        chooseVersion: '选择要降级到的版本（可用：{versions}）', added: '已添加 {id}', addFailed: '添加失败：{error}',
        installed: '已安装', latest: '最新', notInstalled: '未安装', updateAvailable: '有新版本', upToDate: '已是最新',
        recorded: '（记录）', recordedTitle: '从记录读取，未实时探测', explain: '说明', explainTitle: '拉取最新仓库说明并生成双语简介',
        install: '安装', updateTo: '更新到 {version}', downgrade: '降级', uninstall: '卸载', remove: '移除',
        phaseResolve: '解析版本…', phaseDownload: '下载中', phaseChecksum: '校验中', phaseInstall: '安装中',
        installDir: '安装目录', tools: '工具数量', configured: '已配置', notConfigured: '未配置',
        anonymousLimit: '未配置（匿名限流 60 次/时）', replaceToken: '输入新 token 以替换', inputToken: '输入 GitHub token',
        saving: '保存中…', save: '保存', saved: '已保存', saveFailed: '保存失败：{error}', language: '界面语言',
        languageHint: '选择后立即生效，并保存到本机配置。', languageAuto: '跟随系统', languageZh: '简体中文', languageEn: 'English',
        deepseekKey: 'DeepSeek API Key（工具说明）', deepseekHint: '点击卡片上的“说明”，会根据仓库最新 README 生成中英双语简介。',
        replaceKey: '输入新 key 以替换', inputKey: '输入 DeepSeek API key', deepseekModel: 'DeepSeek 模型',
        modelHint: '选择或输入模型名，生成工具说明时使用。', logs: '日志', logCount: '{count} 条', noLogs: '暂无日志',
        addDialog: '添加工具', githubQuick: 'GitHub 地址（快速加入）', githubQuickHint: '粘贴仓库地址，自动识别并预填字段，可手动微调',
        githubPlaceholder: 'https://github.com/owner/repo 或 owner/repo', parsing: '解析中…', parse: '解析', inputGithub: '请输入 GitHub 地址',
        parsed: '已识别 {repo}，字段已预填，请确认后添加。', parseFailed: '解析失败：{error}', required: '{field} 不能为空',
        divider: '', id: 'ID *', idHint: '唯一标识，例如 asc', name: '名称', nameHint: '显示名，默认同 ID', repo: 'GitHub 仓库 *',
        binary: '二进制名', binaryHint: '安装后的可执行文件名，默认同 ID', assetPattern: '资产命名模板', assetHint: '{name}/{version}/{os}/{arch} 占位符',
        checksumPattern: '校验文件模板', versionCommand: '版本命令', versionCommandHint: '默认 --version', versionRegex: '版本正则',
        versionRegexHint: '提取版本的正则捕获组，留空自动匹配 x.y.z', platformMap: '平台映射 JSON', platformHint: 'GOOS/GOARCH → asset 命名，留空用默认',
        invalidPlatformMap: 'platform_map 不是合法 JSON', optionalInstallDir: '安装目录（可选）', optionalInstallHint: '留空使用全局安装目录', cancel: '取消', adding: '添加中…', add: '添加',
        details: '{name} · 说明', loadingExplain: '正在拉取最新说明…', releaseNotes: '最新更新说明', versionN: '版本 {number}', noReleaseNotes: '（该版本未提供更新说明）',
    },
    en: {
        toolCount: '{count} tools', refresh: 'Refresh', settings: 'Settings', addTool: '+ Add tool',
        empty: 'No tools are being managed yet', addFirst: 'Add your first tool', operationFailed: 'Operation failed',
        operationSuccess: '{id} {operation} {version} succeeded', confirmUninstall: 'Uninstall {id}?',
        confirmRemove: 'Remove {id} from the list (the installed binary will remain)?', noVersions: '{id}: no versions available',
        chooseVersion: 'Choose a version to install (available: {versions})', added: 'Added {id}', addFailed: 'Could not add tool: {error}',
        installed: 'Installed', latest: 'Latest', notInstalled: 'Not installed', updateAvailable: 'Update available', upToDate: 'Up to date',
        recorded: ' (recorded)', recordedTitle: 'Read from saved state; not detected live', explain: 'About', explainTitle: 'Fetch repository details and generate a bilingual summary',
        install: 'Install', updateTo: 'Update to {version}', downgrade: 'Downgrade', uninstall: 'Uninstall', remove: 'Remove',
        phaseResolve: 'Resolving…', phaseDownload: 'Downloading', phaseChecksum: 'Verifying', phaseInstall: 'Installing',
        installDir: 'Install directory', tools: 'Tools', configured: 'Configured', notConfigured: 'Not configured',
        anonymousLimit: 'Not configured (60 anonymous requests/hour)', replaceToken: 'Enter a new token to replace it', inputToken: 'Enter GitHub token',
        saving: 'Saving…', save: 'Save', saved: 'Saved', saveFailed: 'Could not save: {error}', language: 'Interface language',
        languageHint: 'Changes take effect immediately and are saved locally.', languageAuto: 'System default', languageZh: '简体中文', languageEn: 'English',
        deepseekKey: 'DeepSeek API key (tool summaries)', deepseekHint: 'Use “About” on a tool card to generate a bilingual summary from its latest README.',
        replaceKey: 'Enter a new key to replace it', inputKey: 'Enter DeepSeek API key', deepseekModel: 'DeepSeek model',
        modelHint: 'Choose or enter the model used to generate tool summaries.', logs: 'Logs', logCount: '{count} entries', noLogs: 'No logs yet',
        addDialog: 'Add tool', githubQuick: 'GitHub repository (quick add)', githubQuickHint: 'Paste a repository URL to detect and prefill the fields below',
        githubPlaceholder: 'https://github.com/owner/repo or owner/repo', parsing: 'Parsing…', parse: 'Parse', inputGithub: 'Enter a GitHub repository',
        parsed: 'Detected {repo}. Review the prefilled fields before adding.', parseFailed: 'Could not parse repository: {error}', required: '{field} is required',
        id: 'ID *', idHint: 'Unique identifier, e.g. asc', name: 'Name', nameHint: 'Display name; defaults to ID', repo: 'GitHub repository *',
        binary: 'Binary name', binaryHint: 'Installed executable name; defaults to ID', assetPattern: 'Asset pattern', assetHint: '{name}/{version}/{os}/{arch} placeholders',
        checksumPattern: 'Checksum file pattern', versionCommand: 'Version command', versionCommandHint: 'Defaults to --version', versionRegex: 'Version regex',
        versionRegexHint: 'Capture the version; leave blank to detect x.y.z', platformMap: 'Platform map JSON', platformHint: 'Map GOOS/GOARCH to asset names; leave blank for defaults',
        invalidPlatformMap: 'platform_map must be valid JSON', optionalInstallDir: 'Install directory (optional)', optionalInstallHint: 'Leave blank to use the global directory', cancel: 'Cancel', adding: 'Adding…', add: 'Add',
        details: '{name} · About', loadingExplain: 'Fetching the latest details…', releaseNotes: 'Latest release notes', versionN: 'Version {number}', noReleaseNotes: '(No release notes provided)',
    },
};

export function resolvedLanguage(preference = 'auto') {
    if (preference === 'zh-CN' || preference === 'en') return preference;
    return (navigator.language || '').toLowerCase().startsWith('zh') ? 'zh-CN' : 'en';
}

export function createTranslator(preference) {
    const locale = resolvedLanguage(preference);
    return (key, values = {}) => {
        let value = messages[locale][key] ?? messages.en[key] ?? key;
        Object.entries(values).forEach(([name, replacement]) => {
            value = value.replaceAll(`{${name}}`, String(replacement));
        });
        return value;
    };
}
