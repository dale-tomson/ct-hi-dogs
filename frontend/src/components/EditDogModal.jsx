import { useState } from "react";

export default function EditDogModal({ dog, onClose, onUpdated, apiFetch }) {
  const [subBreeds, setSubBreeds] = useState([...dog.sub_breeds]);
  const [newSubInput, setNewSubInput] = useState("");
  const [error, setError] = useState(null);
  const [submitting, setSub] = useState(false);

  const handleAddSubBreed = () => {
    if (!newSubInput.trim()) return;
    const lower = newSubInput.trim().toLowerCase();
    
    if (!/^[a-z]+$/.test(lower)) {
      setError(`"${lower}" must contain only lowercase letters a–z.`);
      return;
    }
    
    if (subBreeds.includes(lower)) {
      setError(`"${lower}" is already in the list.`);
      return;
    }
    
    setSubBreeds([...subBreeds, lower]);
    setNewSubInput("");
    setError(null);
  };

  const handleRemoveSubBreed = (index) => {
    setSubBreeds(subBreeds.filter((_, i) => i !== index));
    setError(null);
  };

  const handleSubmit = async () => {
    setError(null);
    setSub(true);
    try {
      const res = await apiFetch(`/api/dogs/${dog.breed}`, {
        method: "PUT",
        body: JSON.stringify({ sub_breeds: subBreeds }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data.error || `Error ${res.status}`);
      onUpdated();
      onClose();
    } catch (err) {
      setError(err.message);
    } finally {
      setSub(false);
    }
  };

  const onKeyDown = (e) => {
    if (e.key === "Escape") onClose();
  };

  const onInputKeyDown = (e) => {
    if (e.key === "Enter") {
      e.preventDefault();
      handleAddSubBreed();
    }
  };

  return (
    <div
      className="modal-overlay"
      onClick={onClose}
      role="dialog"
      aria-modal="true"
      aria-label={`Edit ${dog.breed}`}
    >
      <div
        className="modal"
        onClick={(e) => e.stopPropagation()}
        onKeyDown={onKeyDown}
      >
        <h2>
          Edit: <span className="modal-breed-name">{dog.breed}</span>
        </h2>
        <p className="modal-hint">
          Only sub-breeds can be changed here. To rename a breed, delete it and
          create a new one.
        </p>

        {error && (
          <p className="modal-error" role="alert">
            {error}
          </p>
        )}

        {/* SECTION 1: Add New Sub-Breed */}
        <div className="modal-section">
          <h3 className="section-title">Add New Sub-Breed</h3>
          <div className="input-group">
            <input
              id="new-sub-breed"
              type="text"
              value={newSubInput}
              onChange={(e) => setNewSubInput(e.target.value)}
              onKeyDown={onInputKeyDown}
              placeholder="e.g., labrador"
              autoFocus
              disabled={submitting}
            />
            <button
              className="btn btn-small"
              onClick={handleAddSubBreed}
              disabled={submitting || !newSubInput.trim()}
            >
              Add
            </button>
          </div>
        </div>

        {/* SECTION 2: Sub-Breeds List */}
        <div className="modal-section">
          <h3 className="section-title">
            Sub-Breeds ({subBreeds.length})
          </h3>
          <div className="sub-breeds-list">
            {subBreeds.length > 0 ? (
              <ul>
                {subBreeds.map((sb, idx) => (
                  <li key={idx} className="sub-breed-item">
                    <span>{sb}</span>
                    <button
                      className="btn-delete"
                      onClick={() => handleRemoveSubBreed(idx)}
                      disabled={submitting}
                      aria-label={`Delete ${sb}`}
                    >
                      <svg
                        width="16"
                        height="16"
                        viewBox="0 0 16 16"
                        fill="none"
                        stroke="currentColor"
                        strokeWidth="2"
                      >
                        <line x1="2" y1="2" x2="14" y2="14" />
                        <line x1="14" y1="2" x2="2" y2="14" />
                      </svg>
                    </button>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="empty-message">No sub-breeds yet</p>
            )}
          </div>
        </div>

        <div className="modal-actions">
          <button
            className="btn btn-primary"
            onClick={handleSubmit}
            disabled={submitting}
          >
            {submitting ? "Saving..." : "Save"}
          </button>
          <button
            className="btn btn-secondary"
            onClick={onClose}
            disabled={submitting}
          >
            Cancel
          </button>
        </div>
      </div>
    </div>
  );
}
