import React from "react";
import Button from "../../ui/Button";
import { useTheme } from "./ThemeProvider";
import Sun from "../../ui/Sun";
import Moon from "../../ui/Moon";

function ThemeSelect() {
  const { theme, handleSetTheme } = useTheme();

  return (
    <Button size="large" onClick={handleSetTheme}>
      {theme === "light" ? (
        <Sun className="size-12" />
      ) : (
        <Moon className="size-12" />
      )}
    </Button>
  );
}

export default ThemeSelect;
