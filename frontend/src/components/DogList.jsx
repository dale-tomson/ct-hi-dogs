import DogCard from "./DogCard";

export default function DogList({ dogs, onRefresh, apiFetch }) {
  return (
    <div className="dog-grid">
      {dogs.map((dog) => (
        <DogCard
          key={dog.id}
          dog={dog}
          onRefresh={onRefresh}
          apiFetch={apiFetch}
        />
      ))}
    </div>
  );
}
