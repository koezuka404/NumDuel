import type { ReactNode } from 'react';

type Props = {
  label: string;
  type?: string;
  value: string;
  onChange: (value: string) => void;
  required?: boolean;
};

export default function FormField({ label, type = 'text', value, onChange, required }: Props) {
  return (
    <label>
      {label}
      <input type={type} value={value} onChange={(event) => onChange(event.target.value)} required={required} />
    </label>
  );
}

type StatProps = {
  label: string;
  value: ReactNode;
  valueClassName?: string;
};

export function ProfileStat({ label, value, valueClassName }: StatProps) {
  return (
    <p className="profile-stat">
      <strong>{label}</strong>
      {valueClassName ? <span className={valueClassName}>{value}</span> : value}
    </p>
  );
}

export function FormError({ message }: { message: string }) {
  return <p className="form-error">{message}</p>;
}

export function Spinner({ label, className }: { label?: string; className?: string }) {
  return <div className={className ? `spinner ${className}` : 'spinner'} aria-label={label ?? '読み込み中'} />;
}

type PageHeaderProps = {
  title: string;
  action?: ReactNode;
};

export function PageHeader({ title, action }: PageHeaderProps) {
  if (action) {
    return (
      <div className="page-title-row">
        <h1>{title}</h1>
        {action}
      </div>
    );
  }
  return <h1>{title}</h1>;
}
