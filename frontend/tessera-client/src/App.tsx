import { useState } from 'react';
import LoginPage from './LoginPage';
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

  return (
    <div style={{ padding: '20px' }}>
      <h1>Tessera Dashboard</h1>
      <p>Token: {token.slice(0, 20)}...</p>
      <button onClick={handleLogout}>Logout</button>
      {/* Tomorrow: Add config management here */}
    </div>
  );
}

export default App;