import { useState } from "react";
import Modal from "../../ui/Modal";
import { createPortal } from "react-dom";
import { useTheme } from "../theme-select/ThemeProvider";
import { useSearch } from "./SearchProvider";
import Subheading from "./Subheading";
import ResaultHeading from "./ResaultHeading";
import ResultModal from "./ResultModal";
import Spinner from "../../ui/Spinner";
import Button from "../../ui/Button";

function ResultCard() {
  const { result, isLoading, isError, lastSearch } = useSearch();
  const { content } = result;

  const [isOpenModal, setIsOpenModal] = useState(false);

  function handleSetModal(event) {
    if (!isOpenModal) event.stopPropagation();
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
            <ResultModal
              heading={content?.title}
              content={content?.longresponse}
            />
            {/* <ResultModal content={longresponse} /> */}
          </Modal>,
          document.getElementById("root")
        )}

      <Subheading className="border-b border-gray-700 mb-2 pb-2">
        UPIT: {lastSearch}
      </Subheading>
      {isLoading && <Spinner />}

      {!isLoading && (
        <Subheading>
          REZULTAT - Kliknite bilo gdje unutar sive površine za odlazak na
          stranicu
        </Subheading>
      )}

      <ResaultHeading>{content?.title}</ResaultHeading>
      <p className="text-xl mp:text-lg mb-4">{content?.shortresponse}</p>

      {!isLoading && (
        <Button type="large" onClick={handleSetModal} className="w-full">
          Otvorite Rezultat
        </Button>
      )}
    </div>
  );
}

export default ResultCard;
