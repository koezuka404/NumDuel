type Props = {
  busy: boolean;
  onRebuild: () => void;
};

export default function AdminRankingTab({ busy, onRebuild }: Props) {
  return (
    <section className="admin-section status-panel">
      <p className="status-panel__message">勝利数からランキングを再構築します</p>
      <button type="button" className="btn-primary" onClick={onRebuild} disabled={busy}>
        ランキング再集計
      </button>
    </section>
  );
}
