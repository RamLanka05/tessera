import { useState } from 'react';
import LoginPage from './LoginPage';
import Dashboard from './Dashboard';
import './App.css';

function App() {
  const [token, setToken] = useState<string | null>(localStorage.getItem('token') || null);

  const handleLogin = (newToken: string) => {
    setToken(newToken);
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    setToken(null);
  };

  if (!token) {
    return <LoginPage onLogin={handleLogin} />;
  }

  return <Dashboard token={token} onLogout={handleLogout} />;
}

export default App;