type Props = {
  updatedAt: string | null;
};

export default function RankingFooter({ updatedAt }: Props) {
  return (
    <footer className="ranking-footer">
      <p className="muted">順位は定期更新のため、勝利直後は反映されない場合があります</p>
      {updatedAt && <p>最終更新: {updatedAt}</p>}
    </footer>
  );
}
