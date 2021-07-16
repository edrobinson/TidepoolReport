//Tidepool Report JS Procedures

function validateInputs(){
    alert('In validateInputs()');
    if ($('#usermail').val() == ''or $(#password).val() == ''){
        alert("Email and Password are required. Try again...");
        return false;
    }else{
        return true;
    }
}        
