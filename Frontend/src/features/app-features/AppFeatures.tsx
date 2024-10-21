import { useEffect, useState } from "react";
import Button from "../../ui/Button";
import Cross from "../../ui/Cross";
import Magnifyer from "../../ui/Magnifyer";
import Moon from "../../ui/Moon";
import Sun from "../../ui/Sun";
import { useTheme } from "../theme-select/ThemeProvider";
import ThemeSelect from "../theme-select/ThemeSelect";
import { useZoom } from "./useZoom";

function AppFeatures() {
  const [buttonsVisible, setButtonsVisible] = useState(true);

  const [zoomLevel, handleZoomIn, handleZoomOut] = useZoom();

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

  return (
    <>
      <div className="flex gap-4 justify-center mb-4 mp:flex-col">
        <Button
          size="large"
          onClick={handleZoomOut}
          disabled={zoomLevel === 0.9}
        >
          Odaljite
          <Magnifyer className="size-8" />
        </Button>
        <ThemeSelect />
        <Button size="large" onClick={handleZoomIn} disabled={zoomLevel === 2}>
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
