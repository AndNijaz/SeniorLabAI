import Header from "./ui/Header";
import "./index.css";
import Button from "./ui/Button";
import Container from "./ui/Container";
import ResultCard from "./ui/ResultCard";
import BottomLine from "./ui/BottomLine";
import Magnifyer from "./ui/Magnifyer";
import { useState } from "react";
import Cross from "./ui/Cross";
import Search from "./features/Search/Search";
import AppFeatures from "./features/app-features/AppFeatures";

function App() {
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

      <BottomLine />
    </>
  );
}

export default App;
