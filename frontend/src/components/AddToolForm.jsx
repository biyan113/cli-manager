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

function AddToolForm({onClose, onAdd}) {
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
            setError('请输入 GitHub 地址');
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
            setParseMsg(`已识别 ${sug.repo},字段已预填,请确认后添加。`);
        } catch (e) {
            setError(`解析失败: ${e}`);
        } finally {
            setParsing(false);
        }
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');
        if (!form.id.trim() || !form.repo.trim()) {
            setError('id 和 repo 不能为空');
            return;
        }
        // 解析 platform_map JSON(允许为空)
        let platformMap = {};
        if (form.platform_map.trim()) {
            try {
                platformMap = JSON.parse(form.platform_map);
            } catch {
                setError('platform_map 不是合法 JSON');
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
                    <h2>添加工具</h2>
                    <button className="btn btn-ghost btn-sm" onClick={onClose}>✕</button>
                </div>
                <form onSubmit={handleSubmit} className="modal-body">
                    <div className="quick-add">
                        <Field label="GitHub 地址(快速加入)"
                               hint="粘贴仓库地址,自动识别并预填下方字段,可手动微调">
                            <div className="quick-row">
                                <input value={ghUrl} onChange={e => setGhUrl(e.target.value)}
                                       placeholder="https://github.com/owner/repo 或 owner/repo"/>
                                <button type="button" className="btn btn-primary"
                                        onClick={handleParse} disabled={parsing}>
                                    {parsing ? '解析中…' : '解析'}
                                </button>
                            </div>
                            {parseMsg && <div className="form-info quick-msg">{parseMsg}</div>}
                        </Field>
                    </div>
                    <div className="form-divider"/>
                    <div className="field-grid">
                        <Field label="ID *" hint="唯一标识,如 asc">
                            <input value={form.id} onChange={set('id')} placeholder="asc"/>
                        </Field>
                        <Field label="名称" hint="显示名,默认同 ID">
                            <input value={form.name} onChange={set('name')} placeholder="asc"/>
                        </Field>
                        <Field label="GitHub 仓库 *" hint="owner/repo">
                            <input value={form.repo} onChange={set('repo')} placeholder="rorkai/App-Store-Connect-CLI"/>
                        </Field>
                        <Field label="二进制名" hint="安装后的可执行文件名,默认同 ID">
                            <input value={form.binary} onChange={set('binary')} placeholder="asc"/>
                        </Field>
                        <Field label="资产命名模板" hint="{name}/{version}/{os}/{arch} 占位符">
                            <input value={form.asset_pattern} onChange={set('asset_pattern')}/>
                        </Field>
                        <Field label="校验文件模板">
                            <input value={form.checksums_pattern} onChange={set('checksums_pattern')}/>
                        </Field>
                        <Field label="版本命令" hint="默认 --version">
                            <input value={form.version_cmd} onChange={set('version_cmd')} placeholder="--version"/>
                        </Field>
                        <Field label="版本正则" hint="提取版本的正则捕获组,留空自动匹配 x.y.z">
                            <input value={form.version_regex} onChange={set('version_regex')}
                                   placeholder="^([0-9]+\.[0-9]+\.[0-9]+)"/>
                        </Field>
                    </div>
                    <Field label="平台映射 JSON" hint="GOOS/GOARCH → asset 命名,留空用默认">
                        <textarea value={form.platform_map} onChange={set('platform_map')} rows={3}
                                  className="mono"/>
                    </Field>
                    <Field label="安装目录(可选)" hint="留空用全局 ~/.local/bin">
                        <input value={form.install_dir} onChange={set('install_dir')}
                               placeholder="~/.local/bin"/>
                    </Field>

                    {error && <div className="form-error">{error}</div>}
                    <div className="modal-footer">
                        <button type="button" className="btn btn-ghost" onClick={onClose}>取消</button>
                        <button type="submit" className="btn btn-primary" disabled={saving}>
                            {saving ? '添加中…' : '添加'}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
}

export default AddToolForm;
