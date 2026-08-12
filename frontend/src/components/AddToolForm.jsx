import {useState} from 'react';
import {ParseGithubRepo} from "../../wailsjs/go/main/App";

const DEFAULT_PLATFORM_MAP = '{"darwin":"macOS","linux":"linux","windows":"Windows","arm64":"arm64","amd64":"x86_64"}';

function Field({label, hint, children}) {
    return (
        <label className="field">
            <span className="field-label">{label}</span>
            {children}
            {hint && <span className="field-hint">{hint}</span>}
        </label>
    );
}

function AddToolForm({onClose, onAdd, t}) {
    const [form, setForm] = useState({
        id: '',
        name: '',
        repo: '',
        binary: '',
        asset_pattern: '{name}_{version}_{os}_{arch}',
        checksums_pattern: '{name}_{version}_checksums.txt',
        version_cmd: '--version',
        version_regex: '',
        platform_map: DEFAULT_PLATFORM_MAP,
        install_dir: '',
    });
    const [error, setError] = useState('');
    const [saving, setSaving] = useState(false);
    const [ghUrl, setGhUrl] = useState('');
    const [parsing, setParsing] = useState(false);
    const [parseMsg, setParseMsg] = useState('');

    const set = (key) => (e) => setForm(prev => ({...prev, [key]: e.target.value}));

    // 快速加入:粘贴 GitHub 地址,自动解析并预填表单
    const handleParse = async () => {
        if (!ghUrl.trim()) {
            setError(t('inputGithub'));
            return;
        }
        setError('');
        setParseMsg('');
        setParsing(true);
        try {
            const sug = await ParseGithubRepo(ghUrl.trim());
            setForm(prev => ({
                ...prev,
                id: sug.id || prev.id,
                name: sug.name || prev.name,
                repo: sug.repo || prev.repo,
                binary: sug.binary || prev.binary,
                asset_pattern: sug.asset_pattern || prev.asset_pattern,
                checksums_pattern: sug.checksums_pattern || prev.checksums_pattern,
            }));
            setParseMsg(t('parsed', {repo: sug.repo}));
        } catch (e) {
            setError(t('parseFailed', {error: e}));
        } finally {
            setParsing(false);
        }
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');
        if (!form.id.trim() || !form.repo.trim()) {
            setError(t('required', {field: 'ID / repo'}));
            return;
        }
        // 解析 platform_map JSON(允许为空)
        let platformMap = {};
        if (form.platform_map.trim()) {
            try {
                platformMap = JSON.parse(form.platform_map);
            } catch {
                setError(t('invalidPlatformMap'));
                return;
            }
        }
        const spec = {
            id: form.id.trim(),
            name: form.name.trim() || form.id.trim(),
            repo: form.repo.trim(),
            binary: form.binary.trim() || form.id.trim(),
            asset_pattern: form.asset_pattern.trim() || '{name}_{version}_{os}_{arch}',
            checksums_pattern: form.checksums_pattern.trim() || '{name}_{version}_checksums.txt',
            version_cmd: form.version_cmd.trim() ? form.version_cmd.trim().split(/\s+/) : ['--version'],
            version_regex: form.version_regex.trim(),
            platform_map: platformMap,
            install_dir: form.install_dir.trim(),
        };
        setSaving(true);
        try {
            await onAdd(spec);
        } finally {
            setSaving(false);
        }
    };

    return (
        <div className="modal-overlay" onClick={onClose}>
            <div className="modal" onClick={e => e.stopPropagation()}>
                <div className="modal-header">
                    <h2>{t('addDialog')}</h2>
                    <button className="btn btn-ghost btn-sm" onClick={onClose}>✕</button>
                </div>
                <form onSubmit={handleSubmit} className="modal-body">
                    <div className="quick-add">
                        <Field label={t('githubQuick')} hint={t('githubQuickHint')}>
                            <div className="quick-row">
                                <input value={ghUrl} onChange={e => setGhUrl(e.target.value)}
                                       placeholder={t('githubPlaceholder')}/>
                                <button type="button" className="btn btn-primary"
                                        onClick={handleParse} disabled={parsing}>
                                    {parsing ? t('parsing') : t('parse')}
                                </button>
                            </div>
                            {parseMsg && <div className="form-info quick-msg">{parseMsg}</div>}
                        </Field>
                    </div>
                    <div className="form-divider"/>
                    <div className="field-grid">
                        <Field label={t('id')} hint={t('idHint')}>
                            <input value={form.id} onChange={set('id')} placeholder="asc"/>
                        </Field>
                        <Field label={t('name')} hint={t('nameHint')}>
                            <input value={form.name} onChange={set('name')} placeholder="asc"/>
                        </Field>
                        <Field label={t('repo')} hint="owner/repo">
                            <input value={form.repo} onChange={set('repo')} placeholder="rorkai/App-Store-Connect-CLI"/>
                        </Field>
                        <Field label={t('binary')} hint={t('binaryHint')}>
                            <input value={form.binary} onChange={set('binary')} placeholder="asc"/>
                        </Field>
                        <Field label={t('assetPattern')} hint={t('assetHint')}>
                            <input value={form.asset_pattern} onChange={set('asset_pattern')}/>
                        </Field>
                        <Field label={t('checksumPattern')}>
                            <input value={form.checksums_pattern} onChange={set('checksums_pattern')}/>
                        </Field>
                        <Field label={t('versionCommand')} hint={t('versionCommandHint')}>
                            <input value={form.version_cmd} onChange={set('version_cmd')} placeholder="--version"/>
                        </Field>
                        <Field label={t('versionRegex')} hint={t('versionRegexHint')}>
                            <input value={form.version_regex} onChange={set('version_regex')}
                                   placeholder="^([0-9]+\.[0-9]+\.[0-9]+)"/>
                        </Field>
                    </div>
                    <Field label={t('platformMap')} hint={t('platformHint')}>
                        <textarea value={form.platform_map} onChange={set('platform_map')} rows={3}
                                  className="mono"/>
                    </Field>
                    <Field label={t('optionalInstallDir')} hint={t('optionalInstallHint')}>
                        <input value={form.install_dir} onChange={set('install_dir')}
                               placeholder="~/.local/bin"/>
                    </Field>

                    {error && <div className="form-error">{error}</div>}
                    <div className="modal-footer">
                        <button type="button" className="btn btn-ghost" onClick={onClose}>{t('cancel')}</button>
                        <button type="submit" className="btn btn-primary" disabled={saving}>
                            {saving ? t('adding') : t('add')}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
}

export default AddToolForm;
