import Button from "../../ui/Button";
import Cross from "../../ui/Cross";

function Search() {
  return (
    <>
      <div className="flex gap-2 items-center mb-2 text-lg">
        <p className="text-xl text-main-800 font-semibold">
          Pitajte što god vas zanima 😀
        </p>
        <Button size="small" color="light">
          OČISTI
          <Cross />
        </Button>
      </div>

      <textarea
        className="w-full bg-neutral-300 rounded-md px-2 py-2 text-2xl mb-4"
        rows={4}
        placeholder="Kako poslati sliku na fejzbuku"
      />

      <Button type="medium" className="w-full mb-8">
        Pretražite
      </Button>
    </>
  );
}

export default Search;
