import {useEffect, useMemo, useRef, useState} from 'react';
import './App.css';
import {
    ListTools,
    RefreshAll,
    AddTool,
    RemoveTool,
    InstallTool,
    UpdateTool,
    DowngradeTool,
    UninstallTool,
    SetToken,
    SetDeepSeekToken,
    SetDeepSeekModel,
    SetLanguage,
    GetConfig,
    GetAvailableVersions,
    GetToolExplanation,
} from "../wailsjs/go/main/App";
import {EventsOn, EventsOff} from "../wailsjs/runtime/runtime";

import ToolCard from './components/ToolCard';
import AddToolForm from './components/AddToolForm';
import SettingsModal from './components/SettingsModal';
import ExplainModal from './components/ExplainModal';
import {createTranslator} from './i18n';

function App() {
    const [tools, setTools] = useState([]);
    const [busy, setBusy] = useState({}); // id -> {op, percent, phase}
    const [logs, setLogs] = useState([]);
    const [showAdd, setShowAdd] = useState(false);
    const [showSettings, setShowSettings] = useState(false);
    const [config, setConfig] = useState({install_dir: '', has_token: false, has_deepseek_token: false, deepseek_model: '', language: 'auto', tool_count: 0});
    const [explain, setExplain] = useState(null); // {id, name, loading, text, textEN, releases, error}
    const [toast, setToast] = useState(null);
    const toastTimer = useRef(null);
    const t = useMemo(() => createTranslator(config.language), [config.language]);
    const tRef = useRef(t);
    useEffect(() => { tRef.current = t; }, [t]);

    // 刷新工具列表
    const refresh = async () => {
        const list = await ListTools();
        setTools(list);
    };

    // 一次性加载:初始数据 + 事件订阅
    useEffect(() => {
        refresh();
        GetConfig().then(setConfig);

        const onProgress = (e) => {
            setBusy(prev => ({
                ...prev,
                [e.tool_id]: {op: e.op, percent: e.percent, phase: e.phase, total: e.total, downloaded: e.downloaded},
            }));
        };
        const onStatus = (e) => {
            // 结束:清 busy + 刷新列表
            setBusy(prev => {
                const next = {...prev};
                delete next[e.tool_id];
                return next;
            });
            refresh();
            if (e.status === 'error') {
                showToast('error', `${e.tool_id}: ${e.message || tRef.current('operationFailed')}`);
            } else if (e.status === 'done') {
                showToast('success', tRef.current('operationSuccess', {id: e.tool_id, operation: e.op, version: e.version}));
            }
        };
        const onLog = (e) => {
            setLogs(prev => [...prev.slice(-199), e]);
        };

        EventsOn('tool:progress', onProgress);
        EventsOn('tool:status', onStatus);
        EventsOn('app:log', onLog);
        return () => {
            EventsOff('tool:progress');
            EventsOff('tool:status');
            EventsOff('app:log');
        };
    }, []);

    // toast 自动消失
    const showToast = (type, message) => {
        setToast({type, message});
        if (toastTimer.current) clearTimeout(toastTimer.current);
        toastTimer.current = setTimeout(() => setToast(null), 3500);
    };

    // 操作回调
    const handleAction = async (op, id) => {
        try {
            if (op === 'install') await InstallTool(id);
            else if (op === 'update') await UpdateTool(id);
            else if (op === 'uninstall') {
                if (window.confirm(t('confirmUninstall', {id}))) await UninstallTool(id);
                return;
            }
            else if (op === 'remove') {
                if (window.confirm(t('confirmRemove', {id}))) {
                    await RemoveTool(id);
                    refresh();
                }
                return;
            }
        } catch (e) {
            showToast('error', `${id}: ${e}`);
        }
    };

    // 降级:弹出版本选择
    const handleDowngrade = async (id) => {
        try {
            const versions = await GetAvailableVersions(id);
            if (!versions || versions.length === 0) {
                showToast('error', t('noVersions', {id}));
                return;
            }
            const chosen = window.prompt(t('chooseVersion', {versions: versions.join(', ')}), versions[0]);
            if (!chosen) return;
            await DowngradeTool(id, chosen);
        } catch (e) {
            showToast('error', `${id}: ${e}`);
        }
    };

    // 添加工具
    const handleAdd = async (spec) => {
        try {
            await AddTool(spec);
            setShowAdd(false);
            showToast('success', t('added', {id: spec.id}));
            refresh();
        } catch (e) {
            showToast('error', t('addFailed', {error: e}));
            throw e;
        }
    };

    // 设置 token
    const handleSetToken = async (token) => {
        await SetToken(token);
        GetConfig().then(setConfig);
    };

    // 设置 DeepSeek key
    const handleSetDeepSeekToken = async (token) => {
        await SetDeepSeekToken(token);
        GetConfig().then(setConfig);
    };

    // 设置 DeepSeek 模型
    const handleSetDeepSeekModel = async (model) => {
        await SetDeepSeekModel(model);
        GetConfig().then(setConfig);
    };

    const handleSetLanguage = async (language) => {
        await SetLanguage(language);
        setConfig(prev => ({...prev, language}));
    };

    // 拉取工具中文说明(含最新更新说明)
    const handleExplain = async (id) => {
        const tool = tools.find(t => t.spec.id === id);
        setExplain({id, name: tool?.spec.name || id, loading: true, text: null, textEN: null, releases: [], error: null});
        try {
            const result = await GetToolExplanation(id);
            setExplain(prev => prev && prev.id === id
                ? {...prev, loading: false, text: result.summary, textEN: result.summary_en || '', releases: result.releases || []}
                : prev);
        } catch (e) {
            setExplain(prev => prev && prev.id === id ? {...prev, loading: false, error: String(e)} : prev);
        }
    };

    return (
        <div id="app">
            <header className="topbar">
                <div className="topbar-left" />{/* 左侧留白,为系统原生红黄绿按钮让位 */}
                <div className="brand">
                    <h1>CLI Box</h1>
                    <span className="tool-count">{t('toolCount', {count: tools.length})}</span>
                </div>
                <div className="topbar-actions">
                    <button className="btn btn-ghost" onClick={refresh}>{t('refresh')}</button>
                    <button className="btn btn-ghost" onClick={() => setShowSettings(true)}>{t('settings')}</button>
                    <button className="btn btn-primary" onClick={() => setShowAdd(true)}>{t('addTool')}</button>
                </div>
            </header>

            {toast && <div className={`toast toast-${toast.type}`}>{toast.message}</div>}

            <main className="content">
                {tools.length === 0 ? (
                    <div className="empty-state">
                        <p>{t('empty')}</p>
                        <button className="btn btn-primary" onClick={() => setShowAdd(true)}>{t('addFirst')}</button>
                    </div>
                ) : (
                    <div className="tool-grid">
                        {tools.map(toolItem => (
                            <ToolCard
                                key={toolItem.spec.id}
                                tool={toolItem}
                                busy={busy[toolItem.spec.id]}
                                onAction={handleAction}
                                onDowngrade={handleDowngrade}
                                onExplain={handleExplain}
                                t={t}
                            />
                        ))}
                    </div>
                )}
            </main>

            {showAdd && (
                <AddToolForm
                    onClose={() => setShowAdd(false)}
                    onAdd={handleAdd}
                    t={t}
                />
            )}
            {showSettings && (
                <SettingsModal
                    config={config}
                    hasToken={config.has_token}
                    hasDeepSeekToken={config.has_deepseek_token}
                    deepseekModel={config.deepseek_model}
                    logs={logs}
                    onClose={() => setShowSettings(false)}
                    onSave={handleSetToken}
                    onSaveDeepSeek={handleSetDeepSeekToken}
                    onSaveDeepSeekModel={handleSetDeepSeekModel}
                    onSaveLanguage={handleSetLanguage}
                    t={t}
                />
            )}

            {explain && (
                <ExplainModal
                    toolName={explain.name}
                    text={explain.text}
                    textEN={explain.textEN}
                    releases={explain.releases || []}
                    loading={explain.loading}
                    error={explain.error}
                    onClose={() => setExplain(null)}
                    t={t}
                />
            )}
        </div>
    );
}

export default App;
