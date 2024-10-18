import React from "react";

function Modal({ onSetModal }: { onSetModal: () => void }) {
  return (
    <div
      className="fixed bg-gray-400/75  top-0 right-0 bottom-0 left-0 z-50 transition opacity-100 duration-300"
      onSetModal={onSetModal}
    >
      <div className="w-9/12 absolute top-1/2 left-1/2 bg-white translate-x-[-50%] translate-y-[-50%] p-4 rounded-md shadow-[0px_7px_29px_0px_rgba(100,100,111,0.2)]">
        12312312
      </div>
    </div>
  );
}

export default Modal;
