import type { Folder } from "../types";

type Props = {
  folders: Folder[];
  folderId?: string;
  onChange: (folderId: string | null) => void;
};

export function FolderPicker({ folders, folderId, onChange }: Props) {
  return (
    <label className="folder-picker">
      <span className="folder-picker-label">Папка</span>
      <select
        className="folder-picker-select"
        value={folderId ?? ""}
        onChange={(event) => {
          const value = event.target.value;
          onChange(value === "" ? null : value);
        }}
      >
        <option value="">Без папки</option>
        {folders.map((folder) => (
          <option key={folder.id} value={folder.id}>
            {folder.name}
          </option>
        ))}
      </select>
    </label>
  );
}
