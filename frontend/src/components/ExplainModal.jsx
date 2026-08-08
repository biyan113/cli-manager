import {useEffect, useState} from 'react';

function ExplainText({text}) {
    // 简单 Markdown 渲染:加粗、行内代码、列表项;其余按原样换行显示。
    const html = text
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
        .replace(/`([^`]+)`/g, '<code>$1</code>')
        .replace(/\n-\s+/g, '\n• ');
    return <div className="explain-text" dangerouslySetInnerHTML={{__html: html}}/>;
}

const formatDate = (iso) => (iso ? iso.slice(0, 10) : '');

function ReleaseList({releases}) {
    if (!releases || releases.length === 0) return null;
    return (
        <div className="release-list">
            <h3 className="release-heading">最新更新说明</h3>
            {releases.map((r, i) => (
                <div className="release-item" key={r.tag_name || `${i}`}>
                    <div className="release-meta">
                        <span className="release-tag mono">{r.tag_name || r.name || `版本 ${i + 1}`}</span>
                        {r.published_at && <span className="release-date">{formatDate(r.published_at)}</span>}
                    </div>
                    {r.body ? <ExplainText text={r.body}/> : <span className="release-empty">(该版本未提供更新说明)</span>}
                </div>
            ))}
        </div>
    );
}

function ExplainModal({toolName, text, textEN, releases, loading, error, onClose}) {
    const [lang, setLang] = useState('zh');

    // 双语内容异步加载完成后,若英文不可用则回退中文。
    useEffect(() => {
        if (!textEN && lang === 'en') setLang('zh');
    }, [textEN, lang]);

    const shown = lang === 'en' && textEN ? textEN : text;

    return (
        <div className="modal-overlay" onClick={onClose}>
            <div className="modal modal-sm" onClick={e => e.stopPropagation()}>
                <div className="modal-header">
                    <h2>{toolName} · 说明</h2>
                    <button className="btn btn-ghost btn-sm" onClick={onClose}>✕</button>
                </div>
                <div className="modal-body">
                    {loading && <div className="explain-loading">正在拉取最新说明…</div>}
                    {error && <div className="form-error">{error}</div>}
                    {textEN && (
                        <div className="explain-lang">
                            <button className={`lang-btn${lang === 'zh' ? ' active' : ''}`}
                                    onClick={() => setLang('zh')}>中文</button>
                            <button className={`lang-btn${lang === 'en' ? ' active' : ''}`}
                                    onClick={() => setLang('en')}>English</button>
                        </div>
                    )}
                    {shown && <ExplainText text={shown}/>}
                    {!loading && shown && <ReleaseList releases={releases}/>}
                </div>
            </div>
        </div>
    );
}

export default ExplainModal;
