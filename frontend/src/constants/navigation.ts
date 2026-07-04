export type NavLink = {
  to: string;
  label: string;
};

export const NAV = {
  matching: [
    { to: '/ranking', label: 'ランキング' },
    { to: '/profile', label: 'プロフィール' },
  ],
  ranking: [
    { to: '/matching', label: 'マッチング' },
    { to: '/profile', label: 'プロフィール' },
  ],
  profile: [
    { to: '/matching', label: 'マッチング' },
    { to: '/ranking', label: 'ランキング' },
  ],
} as const satisfies Record<string, NavLink[]>;
