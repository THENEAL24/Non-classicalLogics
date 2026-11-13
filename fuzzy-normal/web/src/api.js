async function request(path, options) {
  const res = await fetch(path, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(data.error || res.statusText)
  return data
}

export const api = {
  journal: () => request('/api/journal'),
  getSet: () => request('/api/set'),
  generate: (body) => request('/api/set/generate', { method: 'POST', body: JSON.stringify(body) }),
  ops: (body) => request('/api/ops', { method: 'POST', body: JSON.stringify(body) }),
  membership: (body) => request('/api/membership', { method: 'POST', body: JSON.stringify(body) }),
  cluster: (body) => request('/api/cluster', { method: 'POST', body: JSON.stringify(body) }),
  clusterRefine: (body) =>
    request('/api/cluster/refine', { method: 'POST', body: JSON.stringify(body) }),
  intervalAnalyze: (body) =>
    request('/api/intervals/analyze', { method: 'POST', body: JSON.stringify(body) }),
  intervalCluster: (body) =>
    request('/api/intervals/cluster', { method: 'POST', body: JSON.stringify(body) }),
  intervalClusterRefine: (body) =>
    request('/api/intervals/cluster/refine', { method: 'POST', body: JSON.stringify(body) }),
}
