import { useState } from "react";
import EditDogModal from "./EditDogModal";

export default function DogCard({ dog, onRefresh, apiFetch }) {
  const [showEdit, setShowEdit] = useState(false);
  const [confirmDelete, setConfirm] = useState(false);
  const [optimisticHide, setHide] = useState(false); // hide card during delete
  const [deleteError, setDeleteError] = useState(null);

  const handleDelete = async () => {
    setHide(true); // Hide immediately for better UX, will revert if error occurs
    setDeleteError(null);
    try {
      const res = await apiFetch(`/api/dogs/${dog.breed}`, {
        method: "DELETE",
      });
      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data.error || `Error ${res.status}`);
      }
      onRefresh(); // sync parent state
    } catch (err) {
      setHide(false);
      setConfirm(false);
      setDeleteError(err.message);
    }
  };

  if (optimisticHide) return null;

  return (
    <div className="dog-card">
      <div className="dog-card-header">
        <h3 className="dog-card-title">{dog.breed}</h3>
        <div className="dog-card-actions">
          <button
            className="btn btn-edit"
            onClick={() => {
              setShowEdit(true);
              setDeleteError(null);
            }}
            aria-label={`Edit ${dog.breed}`}
          >
            Edit
          </button>
          {confirmDelete ? (
            <>
              <button className="btn btn-danger" onClick={handleDelete}>
                Confirm
              </button>
              <button
                className="btn btn-secondary"
                onClick={() => setConfirm(false)}
              >
                Cancel
              </button>
            </>
          ) : (
            <button
              className="btn btn-danger-outline"
              onClick={() => setConfirm(true)}
              aria-label={`Delete ${dog.breed}`}
            >
              Delete
            </button>
          )}
        </div>
      </div>

      {deleteError && (
        <p className="card-error" role="alert">
          {deleteError}
        </p>
      )}

      <button
        className="sub-breeds-count-btn"
        onClick={() => {
          setShowEdit(true);
          setDeleteError(null);
        }}
        aria-label={`Manage sub-breeds for ${dog.breed}`}
      >
        {dog.sub_breeds.length} sub-breed{dog.sub_breeds.length !== 1 ? "s" : ""}
      </button>

      {showEdit && (
        <EditDogModal
          dog={dog}
          onClose={() => setShowEdit(false)}
          onUpdated={onRefresh}
          apiFetch={apiFetch}
        />
      )}
    </div>
  );
}
