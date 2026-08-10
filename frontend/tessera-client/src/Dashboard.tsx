import { useState } from 'react';

interface DashboardProps {
  token: string;
  onLogout: () => void;
}

export default function Dashboard({ token, onLogout }: DashboardProps) {
  const [configName, setConfigName] = useState('');
  const [currentConfig, setCurrentConfig] = useState<any>(null);
  const [versions, setVersions] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const fetchConfig = async () => {
    if (!configName.trim()) return;
    setLoading(true);
    setError('');
    
    try {
      // Get current config
      const res1 = await fetch(`http://localhost:8080/api/v1/config/${configName}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res1.ok) throw new Error('Config not found');
      const config = await res1.json();
      setCurrentConfig(config);

      // Get versions
      const res2 = await fetch(`http://localhost:8080/api/v1/config/${configName}/versions`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res2.ok) throw new Error('Failed to fetch versions');
      const versionData = await res2.json();
      setVersions(versionData.versions);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Error');
    } finally {
      setLoading(false);
    }
  };

  const rollback = async (versionId: number) => {
    try {
      const res = await fetch(`http://localhost:8080/api/v1/config/${configName}/rollback/${versionId}`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res.ok) throw new Error('Rollback failed');
      alert('Rollback successful!');
      fetchConfig(); // Refresh
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Rollback failed');
    }
  };

  return (
    <div>
      <h1>Tessera Dashboard</h1>
      <button onClick={onLogout}>Logout</button>

      <div>
        <h2>Load Config</h2>
        <input
          value={configName}
          onChange={(e) => setConfigName(e.target.value)}
          placeholder="Enter config name (e.g., test_config)"
        />
        <button onClick={fetchConfig} disabled={loading}>
          {loading ? 'Loading...' : 'Load'}
        </button>
      </div>

      {error && <p style={{ color: 'red' }}>{error}</p>}

      {currentConfig && (
        <div>
          <h3>Current Config: {configName}</h3>
          <pre>{JSON.stringify(currentConfig.config_val, null, 2)}</pre>
        </div>
      )}

      {versions.length > 0 && (
        <div>
          <h3>Version History</h3>
          {versions.map((v: any) => (
            <div key={v.ID} style={{ border: '1px solid #ccc', padding: '10px', margin: '10px 0' }}>
              <p><strong>Version {v.ID}</strong> - {v.Message} ({v.Author})</p>
              <p style={{ fontSize: '12px', color: '#666' }}>{v.Timestamp}</p>
              {v.IsActive ? (
                <span style={{ color: 'green' }}>✓ Active</span>
              ) : (
                <button onClick={() => rollback(v.ID)}>Rollback to v{v.ID}</button>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}