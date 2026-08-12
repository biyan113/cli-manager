import {useState} from 'react';
import LogPanel from './LogPanel';

function SettingsModal({config, hasToken, hasDeepSeekToken, deepseekModel, logs, onClose, onSave, onSaveDeepSeek, onSaveDeepSeekModel, onSaveLanguage, t}) {
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
            setMessage(t('saved'));
        } catch (err) {
            setMessage(t('saveFailed', {error: err}));
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
            setDsMessage(t('saved'));
        } catch (err) {
            setDsMessage(t('saveFailed', {error: err}));
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
            setDsModelMessage(t('saved'));
        } catch (err) {
            setDsModelMessage(t('saveFailed', {error: err}));
        } finally {
            setDsModelSaving(false);
        }
    };

    return (
        <div className="modal-overlay" onClick={onClose}>
            <div className="modal" onClick={e => e.stopPropagation()}>
                <div className="modal-header">
                    <h2>{t('settings')}</h2>
                    <button className="btn btn-ghost btn-sm" onClick={onClose}>✕</button>
                </div>
                <div className="modal-body">
                    <div className="setting-row">
                        <span className="setting-label">{t('installDir')}</span>
                        <span className="setting-value mono">{config.install_dir || '~/.local/bin'}</span>
                    </div>
                    <div className="setting-row">
                        <span className="setting-label">{t('tools')}</span>
                        <span className="setting-value">{config.tool_count}</span>
                    </div>
                    <div className="setting-row">
                        <span className="setting-label">GitHub Token</span>
                        <span className="setting-value">
                            {hasToken ? <span className="badge badge-ok">{t('configured')}</span> :
                                <span className="badge badge-muted">{t('anonymousLimit')}</span>}
                        </span>
                    </div>
                    <form onSubmit={handleSave} className="token-form">
                        <input
                            type="password"
                            value={token}
                            onChange={e => setToken(e.target.value)}
                            placeholder={hasToken ? t('replaceToken') : t('inputToken')}
                            autoComplete="off"
                            className="mono"
                        />
                        <button type="submit" className="btn btn-primary" disabled={saving || !token.trim()}>
                            {saving ? t('saving') : t('save')}
                        </button>
                    </form>
                    {message && <div className="form-info">{message}</div>}

                    <div className="setting-row">
                        <span className="setting-label">{t('deepseekKey')}</span>
                        <span className="setting-value">
                            {hasDeepSeekToken ? <span className="badge badge-ok">{t('configured')}</span> :
                                <span className="badge badge-muted">{t('notConfigured')}</span>}
                        </span>
                    </div>
                    <p className="field-hint">{t('deepseekHint')}</p>
                    <form onSubmit={handleSaveDeepSeek} className="token-form">
                        <input
                            type="password"
                            value={dsToken}
                            onChange={e => setDsToken(e.target.value)}
                            placeholder={hasDeepSeekToken ? t('replaceKey') : t('inputKey')}
                            autoComplete="off"
                            className="mono"
                        />
                        <button type="submit" className="btn btn-primary" disabled={dsSaving || !dsToken.trim()}>
                            {dsSaving ? t('saving') : t('save')}
                        </button>
                    </form>
                    {dsMessage && <div className="form-info">{dsMessage}</div>}

                    <div className="setting-row">
                        <span className="setting-label">{t('deepseekModel')}</span>
                        <span className="setting-value mono">{deepseekModel || 'deepseek-v4-flash'}</span>
                    </div>
                    <p className="field-hint">{t('modelHint')}</p>
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
                            {dsModelSaving ? t('saving') : t('save')}
                        </button>
                    </form>
                    {dsModelMessage && <div className="form-info">{dsModelMessage}</div>}

                    <div className="setting-row">
                        <span className="setting-label">{t('language')}</span>
                        <select className="setting-value" value={config.language || 'auto'}
                                onChange={e => onSaveLanguage(e.target.value)}>
                            <option value="auto">{t('languageAuto')}</option>
                            <option value="zh-CN">{t('languageZh')}</option>
                            <option value="en">{t('languageEn')}</option>
                        </select>
                    </div>
                    <p className="field-hint">{t('languageHint')}</p>

                    <div className="log-section">
                        <div className="log-section-title">{t('logs')}</div>
                        <LogPanel logs={logs} t={t}/>
                    </div>
                </div>
            </div>
        </div>
    );
}

export default SettingsModal;
