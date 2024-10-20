import { useState } from "react";
import Modal from "../../ui/Modal";
import { createPortal } from "react-dom";
import { useTheme } from "../theme-select/ThemeProvider";
import { useSearch } from "./SearchProvider";
import Subheading from "./Subheading";
import ResaultHeading from "./ResaultHeading";
import ResultModal from "./ResultModal";

function ResultCard() {
  const { result } = useSearch();
  const { title, description, link } = result;

  const [isOpenModal, setIsOpenModal] = useState(false);

  function handleSetModal() {
    setIsOpenModal((isOpenModal) => !isOpenModal);
  }

  return (
    <div
      className="w-full rounded-md p-8  bg-neutral-100 mb-16 text-[#333333] dark:bg-gray-800 dark:text-[#b0b0b0ea] mp:p-4"
      onClick={handleSetModal}
    >
      {isOpenModal &&
        createPortal(
          <Modal onSetModal={handleSetModal}>
            <ResultModal heading={title} content={description} pages={link} />
          </Modal>,
          document.getElementById("root")
        )}

      <Subheading>
        REZULTAT - Kliknite bilo gdje unutar sive površine za odlazak na
        stranicu
      </Subheading>
      <ResaultHeading>{title}</ResaultHeading>
      <p className="text-xl mp:text-lg">{description}</p>
    </div>
  );
}

export default ResultCard;
