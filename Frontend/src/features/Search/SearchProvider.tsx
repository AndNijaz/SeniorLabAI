import { createContext, useContext, useEffect, useState } from "react";
import Button from "../../ui/Button";
import { fetchQuery } from "../../services/fetchQuery";

const SearchContext = createContext();

function SearchProvider({ children }) {
  const [search, setSearch] = useState(false);
  const [lastSearch, setLastSearch] = useState(false);
  //
  const [result, setResult] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [isError, setIsError] = useState(false);

  function clearSearch() {
    setSearch(null);
    setResult(false); // Reset result state as well
  }

  useEffect(() => {
    setLastSearch(search);

    const getData = async () => {
      try {
        setIsLoading(true);
        setIsError(false);
        // console.log(search);
        const result = await fetchQuery(search);
        setResult(result);
      } catch (error) {
        setIsError(error.message);
      } finally {
        setIsLoading(false);
      }
    };

    getData();
  }, [search]);

  return (
    <SearchContext.Provider
      value={{
        search,
        setSearch,
        result,
        isError,
        isLoading,
        clearSearch,
        lastSearch,
      }}
    >
      {children}
    </SearchContext.Provider>
  );
}

function useSearch() {
  const context = useContext(SearchContext);
  if (context === undefined)
    throw new Error("PostContext used outside of the Provider.");
  return context;
}

function ClearSearchButton() {
  const { clearSearch } = useSearch();

  return (
    <Button
      type="medium"
      color="light"
      className="w-full mb-32"
      onClick={clearSearch}
    >
      Očistite rezultate
    </Button>
  );
}

export { SearchProvider, useSearch, ClearSearchButton };
