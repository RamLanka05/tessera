import { useEffect, useState } from 'react';
import LoginPage from './LoginPage';
import Dashboard from './Dashboard';

type ThemeMode = 'day' | 'night';

const getInitialTheme = (): ThemeMode => {
  const storedTheme = localStorage.getItem('theme');

  if (storedTheme === 'day' || storedTheme === 'night') {
    return storedTheme;
  }

  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'night' : 'day';
};

function App() {
  const [token, setToken] = useState<string | null>(localStorage.getItem('token') || null);
  const [theme, setTheme] = useState<ThemeMode>(getInitialTheme);

  useEffect(() => {
    const root = document.documentElement;
    root.classList.toggle('dark', theme === 'night');
    localStorage.setItem('theme', theme);
  }, [theme]);

  const handleLogin = (newToken: string) => {
    setToken(newToken);
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    setToken(null);
  };

  const handleThemeToggle = () => {
    setTheme((currentTheme) => (currentTheme === 'day' ? 'night' : 'day'));
  };

  if (!token) {
    return <LoginPage onLogin={handleLogin} theme={theme} onToggleTheme={handleThemeToggle} />;
  }

  return <Dashboard token={token} onLogout={handleLogout} theme={theme} onToggleTheme={handleThemeToggle} />;
}

export default App;