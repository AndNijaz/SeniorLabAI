import React from "react";
import Subheading from "./Subheading";
import ResaultHeading from "./ResaultHeading";
import Button from "../../ui/Button";

function ResultModal({
  heading,
  content,
  pages,
}: {
  heading?: string;
  content: string;
  pages?: [string];
}) {
  return (
    <div>
      <Subheading>REZULTAT - Kliknite X za gašenje prozora</Subheading>

      <ResaultHeading>{heading}</ResaultHeading>

      <hr className="mb-4" />

      <div dangerouslySetInnerHTML={{ __html: content }} />
      {/* <p className="text-xl mp:text-lg text-black/80 px-8 mb-4">{content}</p> */}

      <hr className="mb-4" />
    </div>
  );
}

export default ResultModal;
