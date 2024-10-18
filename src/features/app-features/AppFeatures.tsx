import { useState } from "react";
import Button from "../../ui/Button";
import Cross from "../../ui/Cross";
import Magnifyer from "../../ui/Magnifyer";
import Moon from "../../ui/Moon";
import Sun from "../../ui/Sun";
import { useTheme } from "../theme-select/ThemeProvider";

function AppFeatures() {
  const { theme, handleSetTheme } = useTheme();

  const [buttonsVisible, setButtonsVisible] = useState(true);

  const [zoomLevel, setZoomLevel] = useState(1); // Default scale

  if (!buttonsVisible)
    return (
      <Button
        size="large"
        onClick={() => setButtonsVisible(true)}
        className="mx-auto mb-8"
      >
        Prikažite dugmad
      </Button>
    );

  const handleZoomIn = () => {
    setZoomLevel((prev) => Math.min(prev + 0.1, 2)); // Max zoom level 2x
  };

  const handleZoomOut = () => {
    setZoomLevel((prev) => Math.max(prev - 0.1, 0.5)); // Min zoom level 0.5x
  };

  return (
    <>
      <div className="flex gap-4 justify-center mb-4">
        <Button size="large" onClick={handleZoomOut}>
          Odaljite
          <Magnifyer className="size-8" />
        </Button>
        <Button size="large" onClick={handleSetTheme}>
          {theme === "light" ? (
            <Sun className="size-12" />
          ) : (
            <Moon className="size-12" />
          )}
        </Button>
        <Button size="large" onClick={handleZoomIn}>
          Približite
          <Magnifyer className="size-8" />
        </Button>
      </div>

      <Button
        size="small"
        color="light"
        className="mx-auto mb-6"
        onClick={() => setButtonsVisible(false)}
      >
        Sklonite dugmad
        <Cross />
      </Button>
    </>
  );
}

export default AppFeatures;
