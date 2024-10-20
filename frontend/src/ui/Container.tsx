import React from "react";

function Container({
  children,
  className,
}: {
  children: React.ReactNode;
  className: string;
}) {
  return (
    <div
      className={`w-[896px] mx-auto overflow-scroll no-scrollbar scrollbar-none tt:w-[672px] st:w-[512px] mp:w-[320px] ${className}`}
    >
      {children}
    </div>
  );
}

export default Container;
