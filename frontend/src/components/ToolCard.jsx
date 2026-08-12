function ProgressBar({percent, phase, t}) {
    const phaseLabel = {
        'resolve': t('phaseResolve'),
        'download': t('phaseDownload'),
        'checksum': t('phaseChecksum'),
        'install': t('phaseInstall'),
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

function VersionBadge({installed, updateAvailable, t}) {
    if (!installed) return <span className="badge badge-muted">{t('notInstalled')}</span>;
    if (updateAvailable) return <span className="badge badge-warn">{t('updateAvailable')}</span>;
    return <span className="badge badge-ok">{t('upToDate')}</span>;
}

function ToolCard({tool, busy, onAction, onDowngrade, onExplain, t}) {
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
                        <div className="version-label">{t('installed')}</div>
                        <div className="version-value">
                            {tool.installed ? <strong>{tool.installed_version}</strong> : <span className="muted">—</span>}
                            {tool.installed_from === 'recorded' && <span className="hint" title={t('recordedTitle')}>{t('recorded')}</span>}
                        </div>
                    </div>
                    <div className="version-arrow">→</div>
                    <div className="version-cell">
                        <div className="version-label">{t('latest')}</div>
                        <div className="version-value">
                            {tool.latest_version ? <strong>{tool.latest_version}</strong> : <span className="muted">—</span>}
                        </div>
                    </div>
                    <div className="version-cell version-badge-cell">
                        <VersionBadge
                            installed={tool.installed}
                            latest={tool.latest_version}
                            updateAvailable={tool.update_available}
                            t={t}
                        />
                    </div>
                </div>

                {tool.error && <div className="card-error" title={tool.error}>{tool.error}</div>}

                {isBusy && <ProgressBar percent={busy.percent} phase={busy.phase} t={t}/>}
            </div>

            <div className="card-actions">
                <button className="btn btn-ghost btn-sm" disabled={isBusy}
                        onClick={() => onExplain(spec.id)} title={t('explainTitle')}>
                    {t('explain')}
                </button>
                {canInstall && (
                    <button className="btn btn-primary btn-sm" onClick={() => onAction('install', spec.id)}>
                        {t('install')}
                    </button>
                )}
                {canUpdate && (
                    <button className="btn btn-primary btn-sm" onClick={() => onAction('update', spec.id)}>
                        {t('updateTo', {version: tool.latest_version})}
                    </button>
                )}
                {canDowngrade && !isBusy && (
                    <button className="btn btn-ghost btn-sm" onClick={() => onDowngrade(spec.id)}>
                        {t('downgrade')}
                    </button>
                )}
                {tool.installed && (
                    <button className="btn btn-danger btn-sm" disabled={isBusy}
                            onClick={() => onAction('uninstall', spec.id)}>
                        {t('uninstall')}
                    </button>
                )}
                <button className="btn btn-ghost btn-sm" disabled={isBusy}
                        onClick={() => onAction('remove', spec.id)}>
                    {t('remove')}
                </button>
            </div>
        </div>
    );
}

export default ToolCard;
