const ThemeContext = createContext();

const initialState = {
  theme: "light",
};
function ThemeProvider({ children }) {
  const [theme, setTheme] = useState(initialState.theme);

  function handleSetTheme(theme) {
    setTheme(theme);
  }

  return (
    <ThemeContext.Provider value={{ theme, handleSetTheme }}>
      {children}
    </ThemeContext.Provider>
  );
}

function useTheme() {
  const context = useContext(ThemeContext);
  if (context === undefined)
    throw new Error("PostContext used outside of the Provider.");
  return context;
}

export { ThemeProvider, useTheme };
