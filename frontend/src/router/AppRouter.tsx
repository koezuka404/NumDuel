import { Navigate, Route, Routes } from 'react-router-dom';
import AdminPage from '../pages/AdminPage';
import GamePage from '../pages/GamePage';
import LoginPage from '../pages/LoginPage';
import MatchingPage from '../pages/MatchingPage';
import ProfilePage from '../pages/ProfilePage';
import RankingPage from '../pages/RankingPage';
import RegisterPage from '../pages/RegisterPage';
import { GuestOnly, RequireAuth, RequireMaster, RequireUser } from './guards';

export default function AppRouter() {
  return (
    <Routes>
      <Route
        path="/register"
        element={
          <GuestOnly>
            <RegisterPage />
          </GuestOnly>
        }
      />
      <Route
        path="/login"
        element={
          <GuestOnly>
            <LoginPage />
          </GuestOnly>
        }
      />
      <Route
        path="/matching"
        element={
          <RequireAuth>
            <RequireUser>
              <MatchingPage />
            </RequireUser>
          </RequireAuth>
        }
      />
      <Route
        path="/game/:id"
        element={
          <RequireAuth>
            <RequireUser>
              <GamePage />
            </RequireUser>
          </RequireAuth>
        }
      />
      <Route
        path="/ranking"
        element={
          <RequireAuth>
            <RankingPage />
          </RequireAuth>
        }
      />
      <Route
        path="/profile"
        element={
          <RequireAuth>
            <ProfilePage />
          </RequireAuth>
        }
      />
      <Route
        path="/admin"
        element={
          <RequireAuth>
            <RequireMaster>
              <AdminPage />
            </RequireMaster>
          </RequireAuth>
        }
      />
      <Route path="/" element={<Navigate to="/matching" replace />} />
      <Route path="*" element={<Navigate to="/matching" replace />} />
    </Routes>
  );
}
