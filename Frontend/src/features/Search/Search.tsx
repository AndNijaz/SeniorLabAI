import { useRef } from "react";
import Button from "../../ui/Button";
import Cross from "../../ui/Cross";
import { useSearch } from "./SearchProvider";

function Search() {
  const { search, setSearch, clearSearch } = useSearch();
  const queryRef = useRef(null);

  const { result, isLoading, isError } = useSearch();

  function handleSearch() {
    clearSearch();
    const query = queryRef.current.value;
    if (query.trim()) {
      setSearch(query);
    }
    queryRef.current.value = "";
  }

  return (
    <>
      <div className="flex gap-2 items-center mb-2 text-lg">
        <p className="text-xl text-main-800 font-semibold dark:text-main-300 opacity-90">
          {search ? "Interesentno pitanje 😄" : "Pitajte što god vas zanima 😀"}
        </p>
        {search && (
          <Button size="small" color="light" onClick={clearSearch}>
            OČISTI
            <Cross />
          </Button>
        )}
      </div>

      <textarea
        //disabled={isLoading}
        className="w-full bg-neutral-300 rounded-md px-2 py-2 text-2xl mb-4 dark:bg-gray-800 dark:text-[#ffffff]"
        rows={4}
        placeholder={`${
          isLoading
            ? "Trenutno je zahtjev u obradi, molimo sačekajte. "
            : "Na primjer, ukucajte: Kako poslati sliku na fejzbuku"
        } `}
        ref={queryRef}
      />

      <Button type="medium" className="w-full mb-16" onClick={handleSearch}>
        Pretražite
      </Button>
    </>
  );
}

export default Search;
