import { useState, useEffect, useCallback } from "react";
import DogList from "./components/DogList";
import AddDogModal from "./components/AddDogModal";

const API_KEY = import.meta.env.VITE_API_KEY || "";

export const apiFetch = (path, opts = {}) =>
  fetch(path, {
    ...opts,
    headers: {
      "Content-Type": "application/json",
      "X-API-Key": API_KEY,
      ...(opts.headers || {}),
    },
  });

export default function App() {
  const [dogs, setDogs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showAdd, setShowAdd] = useState(false);
  const [search, setSearch] = useState("");

  const fetchDogs = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiFetch("/api/dogs");
      if (!res.ok) throw new Error(`API error ${res.status}`);
      const data = await res.json();
      setDogs(Array.isArray(data) ? data : []);
    } catch (err) {
      setError(
        "Failed to load breeds. Is the API running and API_KEY correct?",
      );
      console.error("fetchDogs:", err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchDogs();
  }, [fetchDogs]);


  const filtered = dogs.filter((d) =>
    d.breed.toLowerCase().includes(search.toLowerCase()),
  );

  return (
    <div className="app">
      <header className="app-header">
        <h1>🐶 Dog Breeds</h1>
        <div className="header-actions">
          <input
            type="text"
            className="search-input"
            placeholder="Search breeds..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            aria-label="Search breeds"
          />
          <button className="btn btn-primary" onClick={() => setShowAdd(true)}>
            + Add Breed
          </button>
        </div>
      </header>

      <main>
        {loading && <p className="state-message">Loading breeds...</p>}

        {!loading && error && (
          <div className="error-banner">
            <p>{error}</p>
            <button className="btn btn-secondary" onClick={fetchDogs}>
              Retry
            </button>
          </div>
        )}

        {!loading && !error && filtered.length === 0 && (
          <div className="empty-state">
            <p>
              {search
                ? `No breeds matching "${search}"`
                : "No breeds yet. Add the first one!"}
            </p>
            {!search && (
              <button
                className="btn btn-primary"
                onClick={() => setShowAdd(true)}
              >
                Add a breed
              </button>
            )}
          </div>
        )}

        {!loading && !error && filtered.length > 0 && (
          <DogList dogs={filtered} onRefresh={fetchDogs} apiFetch={apiFetch} />
        )}
      </main>

      {showAdd && (
        <AddDogModal
          onClose={() => setShowAdd(false)}
          onCreated={fetchDogs}
          apiFetch={apiFetch}
        />
      )}
    </div>
  );
}
