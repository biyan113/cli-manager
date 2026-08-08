function ProgressBar({percent, phase}) {
    const phaseLabel = {
        'resolve': '解析版本…',
        'download': '下载中',
        'checksum': '校验中',
        'install': '安装中',
    };
    return (
        <div className="progress-wrap">
            <div className="progress-bar">
                <div className="progress-fill" style={{width: `${Math.min(100, percent || 0)}%`}}/>
            </div>
            <span className="progress-label">
                {phaseLabel[phase] || phase} {percent > 0 ? `${Math.round(percent)}%` : ''}
            </span>
        </div>
    );
}

function VersionBadge({installed, latest, updateAvailable}) {
    if (!installed) return <span className="badge badge-muted">未安装</span>;
    if (updateAvailable) return <span className="badge badge-warn">有新版本</span>;
    return <span className="badge badge-ok">已是最新</span>;
}

function ToolCard({tool, busy, onAction, onDowngrade, onExplain}) {
    const {spec} = tool;
    const isBusy = !!busy;
    const canInstall = !isBusy && !tool.installed;
    const canUpdate = !isBusy && tool.installed && tool.update_available;
    const canDowngrade = !isBusy && tool.installed;

    return (
        <div className={`card ${isBusy ? 'card-busy' : ''}`}>
            <div className="card-header">
                <div className="card-title">
                    <span className="card-name">{spec.name || spec.id}</span>
                    <span className="card-id">@{spec.id}</span>
                </div>
                <a className="repo-link" href={`https://github.com/${spec.repo}`} target="_blank" rel="noreferrer">
                    {spec.repo}
                </a>
            </div>

            <div className="card-body">
                <div className="version-row">
                    <div className="version-cell">
                        <div className="version-label">已安装</div>
                        <div className="version-value">
                            {tool.installed ? <strong>{tool.installed_version}</strong> : <span className="muted">—</span>}
                            {tool.installed_from === 'recorded' && <span className="hint" title="从记录读取,未实时探测">(记录)</span>}
                        </div>
                    </div>
                    <div className="version-arrow">→</div>
                    <div className="version-cell">
                        <div className="version-label">最新</div>
                        <div className="version-value">
                            {tool.latest_version ? <strong>{tool.latest_version}</strong> : <span className="muted">—</span>}
                        </div>
                    </div>
                    <div className="version-cell version-badge-cell">
                        <VersionBadge
                            installed={tool.installed}
                            latest={tool.latest_version}
                            updateAvailable={tool.update_available}
                        />
                    </div>
                </div>

                {tool.error && <div className="card-error" title={tool.error}>{tool.error}</div>}

                {isBusy && <ProgressBar percent={busy.percent} phase={busy.phase}/>}
            </div>

            <div className="card-actions">
                <button className="btn btn-ghost btn-sm" disabled={isBusy}
                        onClick={() => onExplain(spec.id)} title="拉取最新仓库说明并生成中文简介">
                    说明
                </button>
                {canInstall && (
                    <button className="btn btn-primary btn-sm" onClick={() => onAction('install', spec.id)}>
                        安装
                    </button>
                )}
                {canUpdate && (
                    <button className="btn btn-primary btn-sm" onClick={() => onAction('update', spec.id)}>
                        更新到 {tool.latest_version}
                    </button>
                )}
                {canDowngrade && !isBusy && (
                    <button className="btn btn-ghost btn-sm" onClick={() => onDowngrade(spec.id)}>
                        降级
                    </button>
                )}
                {tool.installed && (
                    <button className="btn btn-danger btn-sm" disabled={isBusy}
                            onClick={() => onAction('uninstall', spec.id)}>
                        卸载
                    </button>
                )}
                <button className="btn btn-ghost btn-sm" disabled={isBusy}
                        onClick={() => onAction('remove', spec.id)}>
                    移除
                </button>
            </div>
        </div>
    );
}

export default ToolCard;
