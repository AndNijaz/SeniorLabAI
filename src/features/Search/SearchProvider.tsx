import { createContext, useContext, useEffect, useState } from "react";
import Button from "../../ui/Button";

const SearchContext = createContext();

function SearchProvider({ children }) {
  const [search, setSearch] = useState(false);
  const [result, setResult] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [isError, setIsError] = useState(false);

  function clearSearch() {
    setSearch(null);
    setResult(false); // Reset result state as well
  }

  useEffect(() => {
    console.log("eddie hall");

    async function fetchSearch() {
      try {
        setIsLoading(true);
        setIsError(false);
        const res = await fetch(`url/${search}`);
        if (!res.ok) throw new Error("Something went bad");
        const data = await res.json();
        setIsLoading(false);
        // setResult(data);
        setResult({
          title: "Dijeljenje fotografija na Facebooku",
          description:
            search +
            " " +
            "Lorem ipsum dolor sit amet, consectetur adipisicing elit. Quo eligendi, reprehenderit molestias mollitia, distinctio voluptatum saepe facere laborum molestiae doloremque a asperiores minima, quidem dignissimos totam aperiam amet repellendus error? Lorem, ipsum dolor sit amet consectetur adipisicing elit. Quis culpa provident amet inventore architecto vel tenetur placeat et? Quia iusto ad itaque blanditiis, doloremque ullam velit fugit unde! Repudiandae, ipsam? Lorem ipsum dolor sit amet, consectetur adipisicing elit. Quo eligendi, reprehenderit molestias mollitia, distinctio voluptatum saepe facere laborum molestiae doloremque a asperiores minima, quidem dignissimos totam aperiam amet repellendus error? Lorem, ipsum dolor sit amet consectetur adipisicing elit. Quis culpa provident amet inventore architecto vel tenetur placeat et? Quia iusto ad itaque blanditiis, doloremque ullam velit fugit unde! Repudiandae, ipsam? Lorem ipsum dolor sit amet, consectetur adipisicing elit. Quo eligendi, reprehenderit molestias mollitia, distinctio voluptatum saepe facere laborum molestiae doloremque a asperiores minima, quidem dignissimos totam aperiam amet repellendus error? Lorem, ipsum dolor sit amet consectetur adipisicing elit. Quis culpa provident amet inventore architecto vel tenetur placeat et? Quia iusto ad itaque blanditiis, doloremque ullam velit fugit unde! Repudiandae, ipsam? Lorem ipsum dolor sit amet, consectetur adipisicing elit. Quo eligendi, reprehenderit molestias mollitia, distinctio voluptatum saepe facere laborum molestiae doloremque a asperiores minima, quidem dignissimos totam aperiam amet repellendus error? Lorem, ipsum dolor sit amet consectetur adipisicing elit. Quis culpa provident amet inventore architecto vel tenetur placeat et? Quia iusto ad itaque blanditiis, doloremque ullam velit fugit unde! Repudiandae, ipsam?",
        });
      } catch (err) {
        setIsLoading(false);
        setIsError(err.message);
      }
    }
    fetchSearch();
    if (search) {
      setResult({
        title: "Dijeljenje fotografija na Facebooku",
        description:
          search +
          " " +
          "Lorem ipsum dolor sit amet, consectetur adipisicing elit. Quo eligendi, reprehenderit molestias mollitia, distinctio voluptatum saepe facere laborum molestiae doloremque a asperiores minima, quidem dignissimos totam aperiam amet repellendus error? Lorem, ipsum dolor sit amet consectetur adipisicing elit. Quis culpa provident amet inventore architecto vel tenetur placeat et? Quia iusto ad itaque blanditiis, doloremque ullam velit fugit unde! Repudiandae, ipsam? Lorem ipsum dolor sit amet, consectetur adipisicing elit. Quo eligendi, reprehenderit molestias mollitia, distinctio voluptatum saepe facere laborum molestiae doloremque a asperiores minima, quidem dignissimos totam aperiam amet repellendus error? Lorem, ipsum dolor sit amet consectetur adipisicing elit. Quis culpa provident amet inventore architecto vel tenetur placeat et? Quia iusto ad itaque blanditiis, doloremque ullam velit fugit unde! Repudiandae, ipsam? Lorem ipsum dolor sit amet, consectetur adipisicing elit. Quo eligendi, reprehenderit molestias mollitia, distinctio voluptatum saepe facere laborum molestiae doloremque a asperiores minima, quidem dignissimos totam aperiam amet repellendus error? Lorem, ipsum dolor sit amet consectetur adipisicing elit. Quis culpa provident amet inventore architecto vel tenetur placeat et? Quia iusto ad itaque blanditiis, doloremque ullam velit fugit unde! Repudiandae, ipsam? Lorem ipsum dolor sit amet, consectetur adipisicing elit. Quo eligendi, reprehenderit molestias mollitia, distinctio voluptatum saepe facere laborum molestiae doloremque a asperiores minima, quidem dignissimos totam aperiam amet repellendus error? Lorem, ipsum dolor sit amet consectetur adipisicing elit. Quis culpa provident amet inventore architecto vel tenetur placeat et? Quia iusto ad itaque blanditiis, doloremque ullam velit fugit unde! Repudiandae, ipsam?",
      });
    }
    console.log("nedzmin");
  }, [search]);

  return (
    <SearchContext.Provider
      value={{ search, setSearch, result, isError, isLoading, clearSearch }}
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
