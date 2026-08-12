import {useEffect, useRef} from 'react';

function LogPanel({logs, t}) {
    const scrollRef = useRef(null);

    // 自动滚动到底部
    useEffect(() => {
        if (scrollRef.current) {
            scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
        }
    }, [logs]);

    const levelClass = (level) => `log-${level || 'info'}`;

    return (
        <div className="log-panel">
            <div className="log-header">
                <span>{t('logs')}</span>
                <span className="muted">{t('logCount', {count: logs.length})}</span>
            </div>
            <div className="log-body" ref={scrollRef}>
                {logs.length === 0 ? (
                    <div className="log-empty">{t('noLogs')}</div>
                ) : (
                    logs.map((log, i) => (
                        <div key={i} className={`log-line ${levelClass(log.level)}`}>
                            <span className="log-time">{log.time ? log.time.slice(11, 19) : ''}</span>
                            <span className="log-msg">{log.message}</span>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
}

export default LogPanel;
