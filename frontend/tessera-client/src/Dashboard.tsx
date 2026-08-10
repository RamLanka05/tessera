import { useMemo, useState, type FormEvent } from 'react';

interface DashboardProps {
  token: string;
  onLogout: () => void;
  theme: 'day' | 'night';
  onToggleTheme: () => void;
}

type ConfigValue = Record<string, unknown> | string | number | boolean | null;

export default function Dashboard({ token, onLogout, theme, onToggleTheme }: DashboardProps) {
  const [configName, setConfigName] = useState('');
  const [currentConfig, setCurrentConfig] = useState<Record<string, unknown> | null>(null);
  const [versions, setVersions] = useState<Array<Record<string, unknown>>>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const apiBase = 'http://localhost:8080/api/v1';

  const currentConfigValue = useMemo<ConfigValue>(() => {
    if (!currentConfig) {
      return null;
    }

    if ('config_val' in currentConfig) {
      return currentConfig.config_val as ConfigValue;
    }

    return currentConfig;
  }, [currentConfig]);

  const fetchConfig = async () => {
    if (!configName.trim()) return;
    setLoading(true);
    setError('');
    
    try {
      const res1 = await fetch(`${apiBase}/config/${configName}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res1.ok) throw new Error('Config not found');
      const config = await res1.json();
      setCurrentConfig(config);

      const res2 = await fetch(`${apiBase}/config/${configName}/versions`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res2.ok) throw new Error('Failed to fetch versions');
      const versionData = await res2.json();
      setVersions(Array.isArray(versionData.versions) ? versionData.versions : []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Error');
    } finally {
      setLoading(false);
    }
  };

  const rollback = async (versionId: number) => {
    try {
      const res = await fetch(`${apiBase}/config/${configName}/rollback/${versionId}`, {
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

  const handleLoadConfig = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    await fetchConfig();
  };

//   const themeLabel = theme === 'day' ? 'Day mode' : 'Night mode';

  return (
    <main className="min-h-screen px-4 py-6 sm:px-6 lg:px-8">
      <div className="mx-auto flex w-full max-w-7xl flex-col gap-6">
        <header className="flex flex-col gap-4 rounded-4xl border border-slate-200/70 bg-white/85 px-6 py-5 shadow-[0_24px_80px_-32px_rgba(15,23,42,0.3)] backdrop-blur-xl dark:border-slate-800/80 dark:bg-slate-950/80 sm:px-8 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <p className="text-sm font-semibold uppercase tracking-[0.3em] text-sky-600 dark:text-sky-400">Tessera</p>
            <h1 className="mt-2 text-3xl font-semibold tracking-tight text-slate-900 dark:text-white">Configuration dashboard</h1>
            <p className="mt-2 max-w-2xl text-sm leading-6 text-slate-600 dark:text-slate-400">
              Inspect the current configuration, review its history, and roll back to a prior version when needed.
            </p>
          </div>

          <div className="flex flex-wrap items-center gap-3">
            <button
              type="button"
              onClick={onToggleTheme}
              aria-label={theme === 'day' ? 'Switch to night mode' : 'Switch to day mode'}
              title={theme === 'day' ? 'Switch to night mode' : 'Switch to day mode'}
              className="inline-flex h-11 w-11 items-center justify-center rounded-full border border-slate-200 bg-white text-slate-700 transition hover:border-slate-300 hover:bg-slate-50 hover:text-slate-900 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-200 dark:hover:border-slate-600 dark:hover:bg-slate-800 dark:hover:text-white"
            >
              {theme === 'day' ? (
                <span aria-hidden="true" className="text-lg leading-none">
                  ☀
                </span>
              ) : (
                <span aria-hidden="true" className="text-lg leading-none">
                  ☾
                </span>
              )}
            </button>
            <button
              onClick={onLogout}
              className="inline-flex items-center rounded-full bg-slate-900 px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-slate-700 dark:bg-sky-500 dark:text-slate-950 dark:hover:bg-sky-400"
            >
              Logout
            </button>
          </div>
        </header>

        <form
          onSubmit={handleLoadConfig}
          className="rounded-4xl border border-slate-200/70 bg-white/90 p-6 shadow-[0_24px_80px_-32px_rgba(15,23,42,0.28)] backdrop-blur-xl dark:border-slate-800/80 dark:bg-slate-950/80 sm:p-8"
        >
          <div className="mb-5 flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
            <div>
              <h2 className="text-xl font-semibold tracking-tight text-slate-900 dark:text-white">Load config</h2>
              <p className="mt-1 text-sm text-slate-600 dark:text-slate-400">
                Enter a config name to fetch the latest value and its version history.
              </p>
            </div>
            <span className="inline-flex w-fit items-center rounded-full bg-sky-50 px-3 py-1.5 text-xs font-semibold text-sky-700 dark:bg-sky-500/10 dark:text-sky-300">
              API authenticated
            </span>
          </div>

          <div className="flex flex-col gap-3 lg:flex-row">
            <div className="flex-1">
              <label className="mb-2 block text-sm font-medium text-slate-700 dark:text-slate-300" htmlFor="config-name">
                Config name
              </label>
              <input
                id="config-name"
                value={configName}
                onChange={(e) => setConfigName(e.target.value)}
                placeholder="Enter config name, for example test_config"
                className="w-full rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-slate-900 outline-none transition placeholder:text-slate-400 focus:border-sky-500 focus:bg-white focus:ring-4 focus:ring-sky-500/10 dark:border-slate-700 dark:bg-slate-900 dark:text-white dark:placeholder:text-slate-500 dark:focus:border-sky-400 dark:focus:bg-slate-950 dark:focus:ring-sky-400/10"
              />
            </div>

            <button
              type="submit"
              disabled={loading}
              className="inline-flex items-center justify-center rounded-2xl bg-slate-900 px-6 py-3 text-sm font-semibold text-white transition hover:bg-slate-700 disabled:cursor-not-allowed disabled:opacity-70 dark:bg-sky-500 dark:text-slate-950 dark:hover:bg-sky-400"
            >
              {loading ? 'Loading...' : 'Load config'}
            </button>
          </div>

          {error && (
            <div className="mt-4 rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700 dark:border-rose-900/70 dark:bg-rose-950/50 dark:text-rose-200">
              {error}
            </div>
          )}
        </form>

        <div className="grid gap-6 xl:grid-cols-[minmax(0,1.25fr)_minmax(360px,0.75fr)]">
          <section className="rounded-4xl border border-slate-200/70 bg-white/90 p-6 shadow-[0_24px_80px_-32px_rgba(15,23,42,0.24)] backdrop-blur-xl dark:border-slate-800/80 dark:bg-slate-950/80 sm:p-8">
            <div className="mb-5 flex items-center justify-between gap-3">
              <div>
                <h2 className="text-xl font-semibold tracking-tight text-slate-900 dark:text-white">Current config</h2>
                <p className="mt-1 text-sm text-slate-600 dark:text-slate-400">Latest configuration snapshot for {configName || 'the selected config'}.</p>
              </div>

              <span className="rounded-full bg-slate-100 px-3 py-1.5 text-xs font-semibold text-slate-700 dark:bg-slate-800 dark:text-slate-200">
                {currentConfig ? 'Loaded' : 'Awaiting input'}
              </span>
            </div>

            {currentConfig ? (
              <div className="space-y-4">
                <div className="grid gap-3 sm:grid-cols-3">
                  <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4 dark:border-slate-800 dark:bg-slate-900">
                    <p className="text-xs font-semibold uppercase tracking-[0.25em] text-slate-500 dark:text-slate-400">Config</p>
                    <p className="mt-2 wrap-break-word text-sm font-semibold text-slate-900 dark:text-white">{configName || 'Unnamed'}</p>
                  </div>
                  <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4 dark:border-slate-800 dark:bg-slate-900">
                    <p className="text-xs font-semibold uppercase tracking-[0.25em] text-slate-500 dark:text-slate-400">Versions</p>
                    <p className="mt-2 text-sm font-semibold text-slate-900 dark:text-white">{versions.length}</p>
                  </div>
                  <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4 dark:border-slate-800 dark:bg-slate-900">
                    <p className="text-xs font-semibold uppercase tracking-[0.25em] text-slate-500 dark:text-slate-400">State</p>
                    <p className="mt-2 text-sm font-semibold text-slate-900 dark:text-white">Active</p>
                  </div>
                </div>

                <div className="overflow-hidden rounded-3xl border border-slate-200 bg-slate-950 text-slate-100 dark:border-slate-800">
                  <div className="flex items-center justify-between border-b border-slate-800 px-4 py-3">
                    <p className="text-sm font-semibold text-slate-200">Configuration payload</p>
                    <p className="text-xs text-slate-400">JSON view</p>
                  </div>
                  <pre className="max-h-130 overflow-auto p-4 text-sm leading-6 text-slate-200">
                    {JSON.stringify(currentConfigValue, null, 2)}
                  </pre>
                </div>
              </div>
            ) : (
              <div className="rounded-3xl border border-dashed border-slate-300 bg-slate-50 p-8 text-sm text-slate-600 dark:border-slate-700 dark:bg-slate-900/60 dark:text-slate-400">
                Load a config to inspect the current JSON payload here.
              </div>
            )}
          </section>

          <section className="rounded-4xl border border-slate-200/70 bg-white/90 p-6 shadow-[0_24px_80px_-32px_rgba(15,23,42,0.24)] backdrop-blur-xl dark:border-slate-800/80 dark:bg-slate-950/80 sm:p-8">
            <div className="mb-5">
              <h2 className="text-xl font-semibold tracking-tight text-slate-900 dark:text-white">Version history</h2>
              <p className="mt-1 text-sm text-slate-600 dark:text-slate-400">Review the latest revisions and roll back when needed.</p>
            </div>

            {versions.length > 0 ? (
              <div className="space-y-3">
                {versions.map((version) => {
                  const versionId = Number(version.ID ?? version.id);
                  const isActive = Boolean(version.IsActive ?? version.isActive);
                  const message = String(version.Message ?? version.message ?? 'No message');
                  const author = String(version.Author ?? version.author ?? 'Unknown');
                  const timestampValue = version.Timestamp ?? version.timestamp;
                  const timestampText = timestampValue ? new Date(String(timestampValue)).toLocaleString() : 'Unknown time';

                  return (
                    <article
                      key={versionId}
                      className="rounded-3xl border border-slate-200 bg-slate-50/80 p-4 transition hover:border-sky-200 hover:bg-slate-50 dark:border-slate-800 dark:bg-slate-900/80 dark:hover:border-sky-700/60"
                    >
                      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                        <div className="space-y-2">
                          <div className="flex flex-wrap items-center gap-2">
                            <span className="rounded-full bg-slate-900 px-3 py-1 text-xs font-semibold text-white dark:bg-sky-500 dark:text-slate-950">
                              Version {versionId}
                            </span>
                            {isActive && (
                              <span className="rounded-full bg-emerald-100 px-3 py-1 text-xs font-semibold text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300">
                                Active
                              </span>
                            )}
                          </div>
                          <p className="text-sm font-medium text-slate-900 dark:text-slate-100">{message}</p>
                          <p className="text-xs text-slate-500 dark:text-slate-400">{author} · {timestampText}</p>
                        </div>

                        {!isActive ? (
                          <button
                            onClick={() => rollback(versionId)}
                            className="inline-flex items-center justify-center rounded-2xl border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700 transition hover:border-sky-200 hover:text-sky-700 dark:border-slate-700 dark:bg-slate-950 dark:text-slate-200 dark:hover:border-sky-700 dark:hover:text-sky-300"
                          >
                            Roll back
                          </button>
                        ) : (
                          <span className="inline-flex items-center rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-2 text-sm font-semibold text-emerald-700 dark:border-emerald-500/20 dark:bg-emerald-500/10 dark:text-emerald-300">
                            Currently live
                          </span>
                        )}
                      </div>
                    </article>
                  );
                })}
              </div>
            ) : (
              <div className="rounded-3xl border border-dashed border-slate-300 bg-slate-50 p-8 text-sm text-slate-600 dark:border-slate-700 dark:bg-slate-900/60 dark:text-slate-400">
                No version history loaded yet.
              </div>
            )}
          </section>
        </div>
      </div>
    </main>
  );
}