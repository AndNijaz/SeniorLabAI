import Header from "./ui/Header";
import "./index.css";
import Button from "./ui/Button";
import Container from "./ui/Container";
import ResultCard from "./ui/ResultCard";
import Footer from "./ui/Footer";
import Search from "./features/Search/Search";
import AppFeatures from "./features/app-features/AppFeatures";
import { useEffect } from "react";
import { useTheme } from "./features/theme-select/ThemeProvider";

function App() {
  // useEffect(,[theme]);

  const x = useTheme();
  console.log(x);

  return (
    <>
      <Header />
      <Container>
        {/* Features Buttons (zoom in, zoom out, theme); Hide Buttons Button */}
        <AppFeatures />

        {/* Text Area Label; Text Area; Clear Button */}
        <Search />

        {/* Resault Caert with its features */}
        <ResultCard />

        <Button type="medium" color="light" className="w-full mb-32">
          Očisti rezultate
        </Button>
      </Container>

      <Footer />
    </>
  );
}

export default App;
