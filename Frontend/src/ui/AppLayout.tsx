import React from "react";
import Header from "./Header";
import Footer from "./Footer";

function AppLayout({ children }) {
  return (
    <div className="flex flex-col h-[100vh]">
      <Header />
      {children}
      <Footer />
    </div>
  );
}

export default AppLayout;
