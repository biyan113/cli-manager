import {useEffect, useRef, useState} from 'react';
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
    GetConfig,
    GetAvailableVersions,
    GetToolExplanation,
} from "../wailsjs/go/main/App";
import {EventsOn, EventsOff} from "../wailsjs/runtime/runtime";

import ToolCard from './components/ToolCard';
import AddToolForm from './components/AddToolForm';
import SettingsModal from './components/SettingsModal';
import ExplainModal from './components/ExplainModal';

function App() {
    const [tools, setTools] = useState([]);
    const [busy, setBusy] = useState({}); // id -> {op, percent, phase}
    const [logs, setLogs] = useState([]);
    const [showAdd, setShowAdd] = useState(false);
    const [showSettings, setShowSettings] = useState(false);
    const [config, setConfig] = useState({install_dir: '', has_token: false, has_deepseek_token: false, deepseek_model: '', tool_count: 0});
    const [explain, setExplain] = useState(null); // {id, name, loading, text, textEN, releases, error}
    const [toast, setToast] = useState(null);
    const toastTimer = useRef(null);

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
                showToast('error', `${e.tool_id}: ${e.message || '操作失败'}`);
            } else if (e.status === 'done') {
                showToast('success', `${e.tool_id} ${e.op} ${e.version} 成功`);
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
                if (window.confirm(`确认卸载 ${id}?`)) await UninstallTool(id);
                return;
            }
            else if (op === 'remove') {
                if (window.confirm(`从清单移除 ${id}(不影响已安装的二进制)?`)) {
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
                showToast('error', `${id}: 没有可用版本`);
                return;
            }
            const chosen = window.prompt(`选择要降级到的版本(可用: ${versions.join(', ')})`, versions[0]);
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
            showToast('success', `已添加 ${spec.id}`);
            refresh();
        } catch (e) {
            showToast('error', `添加失败: ${e}`);
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
                    <h1>CLI Manager</h1>
                    <span className="tool-count">{tools.length} 个工具</span>
                </div>
                <div className="topbar-actions">
                    <button className="btn btn-ghost" onClick={refresh}>刷新</button>
                    <button className="btn btn-ghost" onClick={() => setShowSettings(true)}>设置</button>
                    <button className="btn btn-primary" onClick={() => setShowAdd(true)}>+ 添加工具</button>
                </div>
            </header>

            {toast && <div className={`toast toast-${toast.type}`}>{toast.message}</div>}

            <main className="content">
                {tools.length === 0 ? (
                    <div className="empty-state">
                        <p>还没有管理的工具</p>
                        <button className="btn btn-primary" onClick={() => setShowAdd(true)}>添加第一个工具</button>
                    </div>
                ) : (
                    <div className="tool-grid">
                        {tools.map(t => (
                            <ToolCard
                                key={t.spec.id}
                                tool={t}
                                busy={busy[t.spec.id]}
                                onAction={handleAction}
                                onDowngrade={handleDowngrade}
                                onExplain={handleExplain}
                            />
                        ))}
                    </div>
                )}
            </main>

            {showAdd && (
                <AddToolForm
                    onClose={() => setShowAdd(false)}
                    onAdd={handleAdd}
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
                />
            )}
        </div>
    );
}

export default App;
