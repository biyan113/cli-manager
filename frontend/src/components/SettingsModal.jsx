import {useState} from 'react';

function SettingsModal({config, hasToken, hasDeepSeekToken, deepseekModel, onClose, onSave, onSaveDeepSeek, onSaveDeepSeekModel}) {
    const [token, setToken] = useState('');
    const [saving, setSaving] = useState(false);
    const [message, setMessage] = useState('');
    const [dsToken, setDsToken] = useState('');
    const [dsSaving, setDsSaving] = useState(false);
    const [dsMessage, setDsMessage] = useState('');
    const [dsModel, setDsModel] = useState('');
    const [dsModelSaving, setDsModelSaving] = useState(false);
    const [dsModelMessage, setDsModelMessage] = useState('');

    const handleSave = async (e) => {
        e.preventDefault();
        setSaving(true);
        setMessage('');
        try {
            await onSave(token.trim());
            setToken('');
            setMessage('已保存');
        } catch (err) {
            setMessage(`保存失败: ${err}`);
        } finally {
            setSaving(false);
        }
    };

    const handleSaveDeepSeek = async (e) => {
        e.preventDefault();
        setDsSaving(true);
        setDsMessage('');
        try {
            await onSaveDeepSeek(dsToken.trim());
            setDsToken('');
            setDsMessage('已保存');
        } catch (err) {
            setDsMessage(`保存失败: ${err}`);
        } finally {
            setDsSaving(false);
        }
    };

    const handleSaveDeepSeekModel = async (e) => {
        e.preventDefault();
        setDsModelSaving(true);
        setDsModelMessage('');
        try {
            await onSaveDeepSeekModel(dsModel.trim());
            setDsModel('');
            setDsModelMessage('已保存');
        } catch (err) {
            setDsModelMessage(`保存失败: ${err}`);
        } finally {
            setDsModelSaving(false);
        }
    };

    return (
        <div className="modal-overlay" onClick={onClose}>
            <div className="modal" onClick={e => e.stopPropagation()}>
                <div className="modal-header">
                    <h2>设置</h2>
                    <button className="btn btn-ghost btn-sm" onClick={onClose}>✕</button>
                </div>
                <div className="modal-body">
                    <div className="setting-row">
                        <span className="setting-label">安装目录</span>
                        <span className="setting-value mono">{config.install_dir || '~/.local/bin'}</span>
                    </div>
                    <div className="setting-row">
                        <span className="setting-label">工具数量</span>
                        <span className="setting-value">{config.tool_count}</span>
                    </div>
                    <div className="setting-row">
                        <span className="setting-label">GitHub Token</span>
                        <span className="setting-value">
                            {hasToken ? <span className="badge badge-ok">已配置</span> :
                                <span className="badge badge-muted">未配置(匿名限流 60 次/时)</span>}
                        </span>
                    </div>
                    <form onSubmit={handleSave} className="token-form">
                        <input
                            type="password"
                            value={token}
                            onChange={e => setToken(e.target.value)}
                            placeholder={hasToken ? '输入新 token 以替换' : '输入 GitHub token'}
                            autoComplete="off"
                            className="mono"
                        />
                        <button type="submit" className="btn btn-primary" disabled={saving || !token.trim()}>
                            {saving ? '保存中…' : '保存'}
                        </button>
                    </form>
                    {message && <div className="form-info">{message}</div>}

                    <div className="setting-row">
                        <span className="setting-label">DeepSeek API Key(工具说明)</span>
                        <span className="setting-value">
                            {hasDeepSeekToken ? <span className="badge badge-ok">已配置</span> :
                                <span className="badge badge-muted">未配置</span>}
                        </span>
                    </div>
                    <p className="field-hint">点击卡片上的「说明」按钮,会用 DeepSeek 根据仓库最新 README 生成中文简介。</p>
                    <form onSubmit={handleSaveDeepSeek} className="token-form">
                        <input
                            type="password"
                            value={dsToken}
                            onChange={e => setDsToken(e.target.value)}
                            placeholder={hasDeepSeekToken ? '输入新 key 以替换' : '输入 DeepSeek API key'}
                            autoComplete="off"
                            className="mono"
                        />
                        <button type="submit" className="btn btn-primary" disabled={dsSaving || !dsToken.trim()}>
                            {dsSaving ? '保存中…' : '保存'}
                        </button>
                    </form>
                    {dsMessage && <div className="form-info">{dsMessage}</div>}

                    <div className="setting-row">
                        <span className="setting-label">DeepSeek 模型</span>
                        <span className="setting-value mono">{deepseekModel || 'deepseek-v4-flash'}</span>
                    </div>
                    <p className="field-hint">选择或输入模型名,点击卡片「说明」时按所选模型生成简介。</p>
                    <form onSubmit={handleSaveDeepSeekModel} className="token-form">
                        <input
                            list="ds-models"
                            value={dsModel}
                            onChange={e => setDsModel(e.target.value)}
                            placeholder={deepseekModel || 'deepseek-v4-flash'}
                            autoComplete="off"
                            className="mono"
                        />
                        <datalist id="ds-models">
                            <option value="deepseek-v4-flash"/>
                            <option value="deepseek-v4-pro"/>
                            <option value="deepseek-chat"/>
                            <option value="deepseek-reasoner"/>
                        </datalist>
                        <button type="submit" className="btn btn-primary" disabled={dsModelSaving || !dsModel.trim()}>
                            {dsModelSaving ? '保存中…' : '保存'}
                        </button>
                    </form>
                    {dsModelMessage && <div className="form-info">{dsModelMessage}</div>}
                </div>
            </div>
        </div>
    );
}

export default SettingsModal;
