type TabItem<T extends string> = {
  id: T;
  label: string;
};

type Props<T extends string> = {
  tabs: TabItem<T>[];
  active: T;
  onChange: (id: T) => void;
};

export default function TabBar<T extends string>({ tabs, active, onChange }: Props<T>) {
  return (
    <div className="tabs">
      {tabs.map((tab) => (
        <button
          key={tab.id}
          type="button"
          className={active === tab.id ? 'tab active' : 'tab'}
          onClick={() => onChange(tab.id)}
        >
          {tab.label}
        </button>
      ))}
    </div>
  );
}
