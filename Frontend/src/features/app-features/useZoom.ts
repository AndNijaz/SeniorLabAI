import { useEffect, useState } from "react";

export function useZoom() {
  const [zoomLevel, setZoomLevel] = useState(localStorage.getItem("zoomLevel"));

  if (zoomLevel) console.log("NIJE DOAR NIKAKO");
  else {
    setZoomLevel(1);
    console.log("NE VALJAAAAAAAAAAA");
  }

  useEffect(() => {
    document.body.style.zoom = zoomLevel + "";
  }, [zoomLevel]);

  const handleZoomIn = () => {
    setZoomLevel((prev) => Math.min(prev + 0.1, 2)); // Max zoom level 2x
    localStorage.setItem("zoomLevel", zoomLevel);
  };

  const handleZoomOut = () => {
    setZoomLevel((prev) => Math.max(prev - 0.1, 0.9)); // Min zoom level 0.5x
    localStorage.setItem("zoomLevel", zoomLevel);
  };

  return [zoomLevel, handleZoomIn, handleZoomOut];
}
