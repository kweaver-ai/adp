import intl from 'react-intl-universal';
import { Input, InputProps, Popover } from 'antd';

const InputWithError = ({
  error,
  ...rest
}: InputProps & {
  error?: string;
}) => {
  return error ? (
    <Popover content={error} placement="topLeft">
      <Input placeholder={intl.get('Global.pleaseInput')} status="error" {...rest} />
    </Popover>
  ) : (
    <Input placeholder={intl.get('Global.pleaseInput')} {...rest} />
  );
};

export default InputWithError;
