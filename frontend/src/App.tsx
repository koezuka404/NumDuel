import { BrowserRouter } from 'react-router-dom';
import { AuthProvider } from './hooks/useAuth';
import { ToastProvider } from './hooks/useToast';
import { WebSocketProvider } from './hooks/useWebSocket';
import AppRouter from './router/AppRouter';

export default function App() {
  return (
    <AuthProvider>
      <WebSocketProvider>
        <ToastProvider>
          <BrowserRouter>
            <AppRouter />
          </BrowserRouter>
        </ToastProvider>
      </WebSocketProvider>
    </AuthProvider>
  );
}
