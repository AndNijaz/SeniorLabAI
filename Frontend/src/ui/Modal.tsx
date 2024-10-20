import React from "react";
import Subheading from "../features/Search/Subheading";
import Cross from "./Cross";
import Button from "./Button";

function Modal({
  onSetModal,
  children,
}: {
  onSetModal: () => void;
  heading: string;
  content: string;
  pages: { page: string };
}) {
  return (
    <div
      className="fixed bg-gray-400/75  top-0 right-0 bottom-0 left-0 z-50 transition opacity-100 duration-300"
      onClick={(e) => {
        e.stopPropagation();
        onSetModal();
      }}
    >
      <div
        className="w-[896px] tt:w-[672px] st:w-[512px] mp:w-[320px] absolute top-1/2 left-1/2 bg-white translate-x-[-50%] translate-y-[-50%] p-4 rounded-md shadow-[0px_7px_29px_0px_rgba(100,100,111,0.2)] dark:bg-gray-900"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="relative">
          {children}
          <Button
            type="large"
            className="absolute top-0 right-0 !p-4"
            onClick={onSetModal}
          >
            <Cross className="!size-8" />
          </Button>

          <Button type="large" className="w-full" onClick={onSetModal}>
            Zatvorite Prozor
          </Button>
        </div>
      </div>
    </div>
  );
}

export default Modal;
