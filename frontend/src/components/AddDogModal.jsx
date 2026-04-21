import { useState } from "react";

export default function AddDogModal({ onClose, onCreated, apiFetch }) {
  const [breed, setBreed] = useState("");
  const [subInput, setSubInput] = useState("");
  const [error, setError] = useState(null);
  const [submitting, setSub] = useState(false);

  const handleSubmit = async () => {
    setError(null);

    const breedName = breed.trim().toLowerCase();
    if (!breedName) {
      setError("Breed name is required.");
      return;
    }
    if (!/^[a-z]+$/.test(breedName)) {
      setError(
        "Breed name must contain only lowercase letters a–z (no spaces or numbers).",
      );
      return;
    }

    const subBreeds = subInput
      .split(",")
      .map((s) => s.trim().toLowerCase())
      .filter(Boolean);

    for (const s of subBreeds) {
      if (!/^[a-z]+$/.test(s)) {
        setError(`"${s}" must contain only lowercase letters a–z.`);
        return;
      }
    }

    setSub(true);
    try {
      const res = await apiFetch("/api/dogs", {
        method: "POST",
        body: JSON.stringify({ breed: breedName, sub_breeds: subBreeds }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data.error || `Error ${res.status}`);
      onCreated();
      onClose();
    } catch (err) {
      setError(err.message);
    } finally {
      setSub(false);
    }
  };

  const onKeyDown = (e) => {
    if (e.key === "Enter") handleSubmit();
    if (e.key === "Escape") onClose();
  };

  return (
    <div
      className="modal-overlay"
      onClick={onClose}
      role="dialog"
      aria-modal="true"
      aria-label="Add new breed"
    >
      <div
        className="modal"
        onClick={(e) => e.stopPropagation()}
        onKeyDown={onKeyDown}
      >
        <h2>Add New Breed</h2>

        {error && (
          <p className="modal-error" role="alert">
            {error}
          </p>
        )}

        <label htmlFor="add-breed">Breed Name</label>
        <input
          id="add-breed"
          type="text"
          value={breed}
          onChange={(e) => setBreed(e.target.value)}
          placeholder="e.g. husky"
          autoFocus
          disabled={submitting}
        />

        <label htmlFor="add-subs">
          Sub-breeds{" "}
          <span className="label-hint">(comma-separated, optional)</span>
        </label>
        <input
          id="add-subs"
          type="text"
          value={subInput}
          onChange={(e) => setSubInput(e.target.value)}
          placeholder="e.g. siberian, alaskan"
          disabled={submitting}
        />

        <div className="modal-actions">
          <button
            className="btn btn-primary"
            onClick={handleSubmit}
            disabled={submitting}
          >
            {submitting ? "Creating..." : "Create"}
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
